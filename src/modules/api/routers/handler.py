#!/usr/bin/env python3
from fastapi import APIRouter, HTTPException, Query
from models.worker_result import WorkerResult
from api.worker_handler import WorkerHandler
from api.orchestrator import Orchestrator
from utils.logger import LoggingModule

router = APIRouter()
logger = LoggingModule.get_logger()
worker_handler = WorkerHandler()
orchestrator = Orchestrator()

@router.post("/job/result")
async def handle_job_result(result: WorkerResult, instance_name: str = Query(None)):
    try:
        # Delegate to Orchestrator
        orchestrator.handle_node_completion(result)

        # Cleanup the finished instance
        if instance_name:
            # Use pipeline_id directly from result
            worker_handler.cleanup_worker(str(result.pipeline_id), instance_name)

    except Exception as e:
        logger.error(f"Error handling job result: {e}")
        raise HTTPException(status_code=500, detail="Internal Server Error")