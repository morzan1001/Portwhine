#!/usr/bin/env python3
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, Request

from api.queue import start_queue_thread
from api.docker.container_health import start_container_health_thread

from .routers import trigger, pipeline, handler, worker
from api.docs.main_docs import app_metadata, tags_metadata

# Block logging for specific endpoints
block_endpoints = ["/health"]

class LogFilter(logging.Filter):
    def filter(self, record):
        if record.args and len(record.args) >= 3:
            if record.args[2] in block_endpoints:
                return False
        return True

uvicorn_logger = logging.getLogger("uvicorn.access")
uvicorn_logger.addFilter(LogFilter())

@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup
    queue_thread = start_queue_thread()
    container_health_thread = start_container_health_thread()
    yield
    # Shutdown (if we had logic to stop threads, we would put it here)
    # Currently threads are daemon threads or run forever, so they die with the process.

# FastAPI app with metadata
app = FastAPI(
    title=app_metadata["title"],
    description=app_metadata["description"],
    version=app_metadata["version"],
    license_info=app_metadata["license_info"],
    openapi_tags=tags_metadata,
    lifespan=lifespan
)

# Include routers
app.include_router(trigger.router, prefix="/api/v1", tags=["Trigger"])
app.include_router(pipeline.router, prefix="/api/v1", tags=["Pipelines"])
app.include_router(handler.router, prefix="/api/v1", tags=["Handlers"])
app.include_router(worker.router, prefix="/api/v1", tags=["Worker"])

# Health check endpoint
@app.get("/health", tags=["Health"])
async def health_check():
    return {"status": "healthy"}
