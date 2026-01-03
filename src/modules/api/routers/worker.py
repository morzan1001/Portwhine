#!/usr/bin/env python3
from typing import List, Dict, Any
from fastapi import APIRouter, HTTPException
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule
from models.worker import FFUFWorker, HumbleWorker, ScreenshotWorker, TestSSLWorker, WebAppAnalyzerWorker, NmapWorker, ResolverWorker
from models.responses import NodeConfigExampleResponse
from api.docs.worker_docs import worker_summaries, worker_descriptions
from api.worker_handler import WorkerHandler

router = APIRouter()
logger = LoggingModule.get_logger()

worker_classes = [FFUFWorker, HumbleWorker, ScreenshotWorker, TestSSLWorker, WebAppAnalyzerWorker, NmapWorker, ResolverWorker]

@router.get("/worker", response_model=List[str], summary=worker_summaries["get_workers"], description=worker_descriptions["get_workers"])
async def get_workers():
    return [cls.__name__ for cls in worker_classes]

@router.get(
    "/worker/{name}",
    response_model=NodeConfigExampleResponse,
    summary=worker_summaries["get_worker_config"],
    description=worker_descriptions["get_worker_config"],
)
async def get_worker_config(name: str):
    for cls in worker_classes:
        if cls.__name__ == name:
            # Create example instance with predefined example values
            if cls == NmapWorker:
                example_instance = cls(ports="-p-", arguments="-A")
            elif cls == ResolverWorker:
                example_instance = cls(use_internal=False)
            else:
                example_instance = cls()
            # Clean up the docstring to remove leading/trailing whitespace and newlines
            description = cls.__doc__.strip() if cls.__doc__ else "No description available"
            return NodeConfigExampleResponse(
                description=description,
                example=example_instance.model_dump()
            )
    raise HTTPException(status_code=404, detail="Worker not found")
