#!/usr/bin/env python3
from fastapi import APIRouter, HTTPException
from api.models.pipeline import Pipeline
from api.models.job_payload import JobPayload
from api.worker_handler import WorkerHandler
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule

router = APIRouter()
logger = LoggingModule.get_logger()
es_client = get_elasticsearch_connection()
worker_handler = WorkerHandler()

@router.post("/job/{pipeline_id}/{container_id}")
async def handle_job(pipeline_id: str, container_id: str, payload: JobPayload):
    if not es_client:
        raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
    try:
        # Retrieve the pipeline structure from Elasticsearch
        result = es_client.get(index="pipelines", id=pipeline_id)
        pipeline = Pipeline(**result["_source"])

        # Identify the trigger or workers
        if str(pipeline.trigger._id) == container_id:
            for worker in pipeline.worker:
                worker_handler.start_worker(pipeline_id, worker, payload)
        else:
            # Stop the corresponding Worker container
            #worker = next((w for w in pipeline.worker if str(w.id) == container_id), None)
            #if worker:
            worker_handler.stop_worker(pipeline_id, container_id)
                #worker_handler.start_worker(pipeline_id, worker, payload)
            #else:
            #    raise HTTPException(status_code=404, detail="Worker not found")

    except Exception as e:
        logger.error(f"Error handling job: {e}")
        raise HTTPException(status_code=500, detail="Internal Server Error")