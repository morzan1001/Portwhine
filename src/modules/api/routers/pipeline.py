#!/usr/bin/env python3
import json
import uuid

from elasticsearch import NotFoundError
from fastapi import APIRouter, HTTPException, Query, Request
from typing import List, Dict, Any, Optional

from pydantic import BaseModel, Field

from fastapi.encoders import jsonable_encoder
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule
from utils.helper import strip_runtime_fields
from models.pipeline import Pipeline
from models.responses import (
    PipelineResponse,
    PipelineListItem,
    MessageResponse,
    DeleteResponse,
)
from api.pipeline_handler import PipelineHandler
from models.types import NodeStatus
from api.docs.pipeline_docs import pipeline_summaries, pipeline_descriptions

router = APIRouter()
logger = LoggingModule.get_logger()
es_client = get_elasticsearch_connection()
pipeline_handler = PipelineHandler()


@router.post(
    "/pipeline",
    response_model=PipelineResponse,
    summary=pipeline_summaries["create_pipeline"],
    description=pipeline_descriptions["create_pipeline"],
)
async def create_pipeline(pipeline: Pipeline):
    """
    Create a new pipeline.
    
    Runtime fields (status, inputs, outputs) are not stored - they are computed 
    from class definitions when the pipeline is retrieved.
    """
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        # Store only persistable data using ser_for_storage()
        es_client.index(
            index="pipelines", 
            id=str(pipeline._id), 
            body=pipeline.ser_for_storage(), 
            refresh="wait_for"
        )
        # Return full response with runtime fields
        return pipeline.ser_model()
    except Exception as e:
        logger.error(f"Error creating pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error creating pipeline: {str(e)}")

@router.get(
    "/pipeline/{pipeline_id}",
    response_model=PipelineResponse,
    summary=pipeline_summaries["get_pipeline"],
    description=pipeline_descriptions["get_pipeline"],
)
async def get_pipeline(pipeline_id: str):
    """
    Get a pipeline by ID.
    
    The response includes runtime fields (status, inputs, outputs) that are 
    computed from class definitions, not stored in the database.
    """
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        result = es_client.get(index="pipelines", id=pipeline_id)
        # Strip runtime fields from DB data before loading into model
        clean_data = strip_runtime_fields(result["_source"])
        # Parse the stored data into a Pipeline object to get runtime fields
        pipeline = Pipeline(**clean_data)
        # Return with runtime fields included
        return pipeline.ser_model()
    except NotFoundError:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    except Exception as e:
        logger.error(f"Error retrieving pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error fetching pipeline: {str(e)}")

@router.get(
    "/pipelines",
    response_model=List[PipelineListItem],
    summary=pipeline_summaries["get_all_pipelines"],
    description=pipeline_descriptions["get_all_pipelines"],
)
async def get_all_pipelines(size: int = Query(10, ge=1), page: int = Query(1, ge=1)):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        if not es_client.indices.exists(index="pipelines"):
            return []  # No pipelines index yet - return empty list
        
        from_ = (page - 1) * size
        result = es_client.search(index="pipelines", body={"query": {"match_all": {}}, "from": from_, "size": size}, _source=["id", "name"])
        
        if result["hits"]["total"]["value"] == 0:
            return []  # No pipelines found - return empty list

        pipelines = [{"id": hit["_id"], "name": hit["_source"]["name"], "status": hit["_source"].get("status", "Unknown")} for hit in result["hits"]["hits"]]
        return pipelines
    except HTTPException as http_exc:
        raise http_exc
    except Exception as e:
        logger.error(f"Error retrieving pipelines: {e}")
        raise HTTPException(status_code=500, detail="Error retrieving pipelines")

class PipelinePatch(BaseModel):
    id: uuid.UUID = Field(..., description="Pipeline ID")
    name: Optional[str] = None
    trigger: Optional[Dict[str, Any]] = None
    worker: Optional[List[Dict[str, Any]]] = None
    edges: Optional[List[Dict[str, Any]]] = None


@router.patch(
    "/pipeline",
    response_model=PipelineResponse,
    summary=pipeline_summaries["update_pipeline"],
    description=pipeline_descriptions["update_pipeline"],
)
async def update_pipeline(
    patch: PipelinePatch,
) -> Dict[str, Any]:
    """
    Update a pipeline configuration.
    
    The endpoint accepts a PipelinePatch with trigger/worker/edges data.
    Runtime fields (status, inputs, outputs, instanceHealth) are stripped
    before validation - only persistable fields are accepted.
    
    Temporary IDs starting with 'new-' are remapped to proper UUIDs.
    """
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        # Retrieve the existing pipeline
        pipeline_id = str(patch.id)
        database_output = es_client.get(index="pipelines", id=pipeline_id)["_source"]
        
        # Strip runtime fields from DB data before loading
        clean_db_data = strip_runtime_fields(database_output)
        existing_pipeline: Pipeline = Pipeline(**clean_db_data)
        logger.debug(f"Existing pipeline: {existing_pipeline.name}")

        if existing_pipeline._status == NodeStatus.RUNNING:
            raise HTTPException(status_code=400, detail="Pipeline is running. Stop the pipeline before updating")

        # Update the pipeline using a PATCH model (name/trigger/worker/edges optional)
        updated_fields = patch.model_dump(exclude_unset=True)
        
        # Strip runtime fields from incoming patch data as well
        updated_fields = strip_runtime_fields(updated_fields)

        # Normalize nulls sent by clients
        if "edges" in updated_fields and updated_fields["edges"] is None:
            updated_fields["edges"] = []
        if "worker" in updated_fields and updated_fields["worker"] is None:
            updated_fields["worker"] = []

        # Ensure the ID is preserved as string in the merged payload
        updated_fields["id"] = pipeline_id

        # Merge clean DB data with clean updated fields
        merged_payload = {**clean_db_data, **updated_fields}

        # Helper to check if a string is a valid UUID
        def _is_valid_uuid(val: str) -> bool:
            try:
                uuid.UUID(str(val))
                return True
            except (ValueError, TypeError, AttributeError):
                return False

        # Remap temporary IDs (starting with "new-") to real UUIDs
        id_mapping: Dict[str, str] = {}

        # 1. Check trigger for temp IDs
        trigger_data = merged_payload.get("trigger")
        if isinstance(trigger_data, dict):
            for t_type, t_config in list(trigger_data.items()):
                if isinstance(t_config, dict):
                    tid = t_config.get("id")
                    if tid and (str(tid).startswith("new-") or not _is_valid_uuid(tid)):
                        new_uid = str(uuid.uuid4())
                        id_mapping[tid] = new_uid
                        t_config["id"] = new_uid
                        logger.info(f"Remapped trigger ID {tid} to {new_uid}")

        # 2. Check workers for temp IDs
        workers_data = merged_payload.get("worker")
        if isinstance(workers_data, list):
            for i, w_entry in enumerate(workers_data):
                if isinstance(w_entry, dict):
                    for w_type, w_config in list(w_entry.items()):
                        if isinstance(w_config, dict):
                            wid = w_config.get("id")
                            if wid and (str(wid).startswith("new-") or not _is_valid_uuid(wid)):
                                new_uid = str(uuid.uuid4())
                                id_mapping[wid] = new_uid
                                w_config["id"] = new_uid
                                logger.info(f"Remapped worker {i} ({w_type}) ID {wid} to {new_uid}")

        # 3. Update edges with remapped IDs
        if id_mapping:
            edges_data = merged_payload.get("edges", [])
            for edge in edges_data:
                if isinstance(edge, dict):
                    if edge.get("source") in id_mapping:
                        edge["source"] = id_mapping[edge["source"]]
                    if edge.get("target") in id_mapping:
                        edge["target"] = id_mapping[edge["target"]]

        # Validate and create the updated pipeline
        # Runtime fields have been stripped, so strict validation will pass
        try:
            updated_pipeline: Pipeline = Pipeline(**merged_payload)
        except Exception as e:
            logger.error(f"Pipeline validation failed: {e}")
            raise HTTPException(status_code=422, detail=f"Invalid pipeline data: {str(e)}") from e

        # Save to Elasticsearch using ser_for_storage() to exclude runtime fields
        es_client.index(
            index="pipelines",
            id=pipeline_id,
            body=updated_pipeline.ser_for_storage(),
            refresh="wait_for",
        )
        
        # Return the full response with runtime fields for the frontend
        return updated_pipeline.ser_model()
    except NotFoundError:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    except HTTPException:
        raise
    except Exception as e:
        logger.error(f"Error updating pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error updating pipeline: {str(e)}")

@router.delete(
    "/pipeline/{pipeline_id}",
    response_model=DeleteResponse,
    summary=pipeline_summaries["delete_pipeline"],
    description=pipeline_descriptions["delete_pipeline"],
)
async def delete_pipeline(pipeline_id: str):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        es_client.delete(index="pipelines", id=pipeline_id)
        return DeleteResponse(detail="Pipeline deleted successfully")
    except NotFoundError:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    except Exception as e:
        logger.error(f"Error deleting pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error deleting pipeline: {str(e)}")

@router.post(
    "/pipeline/start/{pipeline_id}",
    response_model=MessageResponse,
    summary=pipeline_summaries["start_pipeline"],
    description=pipeline_descriptions["start_pipeline"],
)
async def start_pipeline(pipeline_id: str):
    pipeline_handler.handle_pipeline_start(pipeline_id)
    return MessageResponse(detail=f"Pipeline {pipeline_id} started successfully.")

@router.post(
    "/pipeline/stop/{pipeline_id}",
    response_model=MessageResponse,
    summary=pipeline_summaries["stop_pipeline"],
    description=pipeline_descriptions["stop_pipeline"],
)
async def stop_pipeline(pipeline_id: str):
    pipeline_handler.handle_pipeline_stop(pipeline_id)
    return MessageResponse(detail=f"Pipeline {pipeline_id} stopped successfully.")

@router.post(
    "/pipeline/cleanup/{pipeline_id}",
    response_model=MessageResponse,
    summary=pipeline_summaries["cleanup_pipeline"],
    description=pipeline_descriptions["cleanup_pipeline"],
)
async def cleanup_containers(pipeline_id: str):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        result = es_client.get(index="pipelines", id=pipeline_id)
        clean_data = strip_runtime_fields(result["_source"])
        pipeline = Pipeline(**clean_data)
        if pipeline._status == NodeStatus.STOPPED:
            pipeline_handler.cleanup_containers(str(pipeline.trigger._id))
            for worker in pipeline.worker:
                pipeline_handler.cleanup_containers(str(worker._id))
                pipeline_handler.reset_instance_count(str(worker._id), pipeline_id)
            return MessageResponse(detail="Containers cleaned up successfully")
        else:
            raise HTTPException(status_code=400, detail="Pipeline is not stopped")
    except NotFoundError:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    except HTTPException as http_exc:
        raise http_exc
    except Exception as e:
        logger.error(f"Error deleting pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error deleting pipeline: {str(e)}")
