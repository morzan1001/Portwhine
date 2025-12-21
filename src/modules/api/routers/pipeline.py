#!/usr/bin/env python3
import json

from elasticsearch import NotFoundError
from fastapi import APIRouter, HTTPException, Query
from typing import List, Dict

from fastapi.encoders import jsonable_encoder
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule
from models.pipeline import Pipeline
from api.pipeline_handler import PipelineHandler
from models.types import NodeStatus
from api.docs.pipeline_docs import pipeline_summaries, pipeline_descriptions

router = APIRouter()
logger = LoggingModule.get_logger()
es_client = get_elasticsearch_connection()
pipeline_handler = PipelineHandler()

@router.post(
    "/pipeline",
    response_model=Pipeline,
    summary=pipeline_summaries["create_pipeline"],
    description=pipeline_descriptions["create_pipeline"],
)
async def create_pipeline(pipeline: Pipeline):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        es_client.index(index="pipelines", id=str(pipeline._id), body=pipeline.ser_model(), refresh="wait_for")
        return pipeline
    except Exception as e:
        logger.error(f"Error creating pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error creating pipeline: {str(e)}")

@router.get(
    "/pipeline/{pipeline_id}",
    response_model=Pipeline,
    summary=pipeline_summaries["get_pipeline"],
    description=pipeline_descriptions["get_pipeline"],
)
async def get_pipeline(pipeline_id: str):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        result = es_client.get(index="pipelines", id=pipeline_id)
        return result["_source"]
    except NotFoundError:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    except Exception as e:
        logger.error(f"Error retrieving pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error fetching pipeline: {str(e)}")

@router.get(
    "/pipelines",
    response_model=List[Dict[str, str]],
    summary=pipeline_summaries["get_all_pipelines"],
    description=pipeline_descriptions["get_all_pipelines"],
)
async def get_all_pipelines(size: int = Query(10, ge=1), page: int = Query(1, ge=1)):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        if not es_client.indices.exists(index="pipelines"):
            raise HTTPException(status_code=404, detail="No pipelines found")
        
        from_ = (page - 1) * size
        result = es_client.search(index="pipelines", body={"query": {"match_all": {}}, "from": from_, "size": size}, _source=["id", "name"])
        
        if result["hits"]["total"]["value"] == 0:
            raise HTTPException(status_code=404, detail="No pipelines found")

        pipelines = [{"id": hit["_id"], "name": hit["_source"]["name"], "status": hit["_source"].get("status", "Unknown")} for hit in result["hits"]["hits"]]
        return pipelines
    except HTTPException as http_exc:
        raise http_exc
    except Exception as e:
        logger.error(f"Error retrieving pipelines: {e}")
        raise HTTPException(status_code=500, detail="Error retrieving pipelines")

@router.patch(
    "/pipeline",
    response_model=Pipeline,
    summary=pipeline_summaries["update_pipeline"],
    description=pipeline_descriptions["update_pipeline"],
)
async def update_pipeline(pipeline: Pipeline) -> Pipeline:
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        # Retrieve the existing pipeline
        database_output = es_client.get(index="pipelines", id=pipeline._id)["_source"]
        existing_pipeline: Pipeline = Pipeline(**database_output)
        logger.debug(f"Existing pipeline data: {existing_pipeline}")

        if existing_pipeline._status == NodeStatus.RUNNING:
            raise HTTPException(status_code=400, detail="Pipeline is running. Stop the pipeline before updating")

        # Update the pipeline
        updated_fields = pipeline.model_dump(exclude_unset=True)
        logger.debug(f"Updated fields: {updated_fields}")
        updated_pipeline: Pipeline = existing_pipeline.model_copy(update=updated_fields)
        logger.debug(f"Updated pipeline data: {updated_pipeline}")

        # Save the updated pipeline to Elasticsearch
        es_client.index(index="pipelines", id=pipeline._id, body=updated_pipeline.model_dump(), refresh="wait_for")
        return updated_pipeline
    except NotFoundError:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    except Exception as e:
        logger.error(f"Error updating pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error updating pipeline: {str(e)}")

@router.delete(
    "/pipeline/{pipeline_id}",
    response_model=dict,
    summary=pipeline_summaries["delete_pipeline"],
    description=pipeline_descriptions["delete_pipeline"],
)
async def delete_pipeline(pipeline_id: str):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        es_client.delete(index="pipelines", id=pipeline_id)
        return {"detail": "Pipeline deleted successfully"}
    except NotFoundError:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    except Exception as e:
        logger.error(f"Error deleting pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error deleting pipeline: {str(e)}")

@router.post(
    "/pipeline/start/{pipeline_id}",
    summary=pipeline_summaries["start_pipeline"],
    description=pipeline_descriptions["start_pipeline"],
)
async def start_pipeline(pipeline_id: str):
    pipeline_handler.handle_pipeline_start(pipeline_id)
    return {"detail": f"Pipeline {pipeline_id} started successfully."}

@router.post(
    "/pipeline/stop/{pipeline_id}",
    summary=pipeline_summaries["stop_pipeline"],
    description=pipeline_descriptions["stop_pipeline"],
)
async def stop_pipeline(pipeline_id: str):
    pipeline_handler.handle_pipeline_stop(pipeline_id)
    return {"detail": f"Pipeline {pipeline_id} stopped successfully."}

@router.post(
    "/pipeline/cleanup/{pipeline_id}",
    summary=pipeline_summaries["cleanup_pipeline"],
    description=pipeline_descriptions["cleanup_pipeline"],
)
async def cleanup_containers(pipeline_id: str):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        result = es_client.get(index="pipelines", id=pipeline_id)
        pipeline = Pipeline(**result["_source"])
        if pipeline._status == NodeStatus.STOPPED:
            pipeline_handler.cleanup_containers(str(pipeline.trigger._id))
            for worker in pipeline.worker:
                pipeline_handler.cleanup_containers(str(worker._id))
                pipeline_handler.reset_instance_count(str(worker._id), pipeline_id)
            return {"detail": "Containers cleaned up successfully"}
        else:
            raise HTTPException(status_code=400, detail="Pipeline is not stopped")
    except NotFoundError:
        raise HTTPException(status_code=404, detail="Pipeline not found")
    except HTTPException as http_exc:
        raise http_exc
    except Exception as e:
        logger.error(f"Error deleting pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error deleting pipeline: {str(e)}")
