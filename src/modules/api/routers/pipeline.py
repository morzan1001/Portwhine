#!/usr/bin/env python3
from fastapi import APIRouter, HTTPException, Query
from typing import List, Dict
from api.models.pipeline import Pipeline
from api.pipeline_handler import PipelineHandler
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule
from api.docs.pipeline_docs import pipeline_responses, pipeline_summaries, pipeline_descriptions

router = APIRouter()
logger = LoggingModule.get_logger()
es_client = get_elasticsearch_connection()
pipeline_handler = PipelineHandler()

@router.post(
    "/pipeline",
    response_model=Pipeline,
    summary=pipeline_summaries["create_pipeline"],
    description=pipeline_descriptions["create_pipeline"],
    responses=pipeline_responses
)
async def create_pipeline(pipeline: Pipeline):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        es_client.index(index="pipelines", id=str(pipeline.id), body=pipeline.ser_model())
        return pipeline
    except Exception as e:
        logger.error(f"Error creating pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error creating pipeline: {str(e)}")

@router.get(
    "/pipeline/{pipeline_id}",
    response_model=Pipeline,
    summary=pipeline_summaries["get_pipeline"],
    description=pipeline_descriptions["get_pipeline"],
    responses=pipeline_responses
)
async def get_pipeline(pipeline_id: str):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        result = es_client.get(index="pipelines", id=pipeline_id)
        return result["_source"]
    except Exception as e:
        logger.error(f"Error retrieving pipeline: {e}")
        raise HTTPException(status_code=404, detail="Pipeline not found")

@router.get(
    "/pipelines",
    response_model=List[Dict[str, str]],
    summary=pipeline_summaries["get_all_pipelines"],
    description=pipeline_descriptions["get_all_pipelines"],
    responses=pipeline_responses
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
        
        pipelines = [{"id": hit["_id"], "name": hit["_source"]["name"]} for hit in result["hits"]["hits"]]
        return pipelines
    except Exception as e:
        logger.error(f"Error retrieving pipelines: {e}")
        raise HTTPException(status_code=500, detail="Error retrieving pipelines")

@router.put(
    "/pipeline/{pipeline_id}",
    response_model=Pipeline,
    summary=pipeline_summaries["update_pipeline"],
    description=pipeline_descriptions["update_pipeline"],
    responses=pipeline_responses
)
async def update_pipeline(pipeline_id: str, pipeline: Pipeline):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        es_client.index(index="pipelines", id=pipeline_id, body=pipeline.ser_model())
        return pipeline
    except Exception as e:
        logger.error(f"Error updating pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error updating pipeline: {str(e)}")

@router.delete(
    "/pipeline/{pipeline_id}",
    response_model=dict,
    summary=pipeline_summaries["delete_pipeline"],
    description=pipeline_descriptions["delete_pipeline"],
    responses=pipeline_responses
)
async def delete_pipeline(pipeline_id: str):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        es_client.delete(index="pipelines", id=pipeline_id)
        return {"message": "Pipeline deleted successfully"}
    except Exception as e:
        logger.error(f"Error deleting pipeline: {e}")
        raise HTTPException(status_code=500, detail=f"Error deleting pipeline: {str(e)}")
    
@router.post(
    "/pipeline/start/{pipeline_id}",
    summary=pipeline_summaries["start_pipeline"],
    description=pipeline_descriptions["start_pipeline"],
    responses=pipeline_responses
)
async def start_pipeline(pipeline_id: str):
    pipeline_handler.handle_pipeline_start(pipeline_id)
    return {"message": f"Pipeline {pipeline_id} started successfully."}

@router.post(
    "/pipeline/stop/{pipeline_id}",
    summary=pipeline_summaries["stop_pipeline"],
    description=pipeline_descriptions["stop_pipeline"],
    responses=pipeline_responses
)
async def stop_pipeline(pipeline_id: str):
    pipeline_handler.handle_pipeline_stop(pipeline_id)
    return {"message": f"Pipeline {pipeline_id} stopped successfully."}