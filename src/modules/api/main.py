#!/usr/bin/env python3
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from pydantic import ValidationError

from api.queue import start_queue_thread
from api.docker.container_health import start_container_health_thread

from .routers import trigger, pipeline, handler, worker, nodes, websocket
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

# Global exception handler for Pydantic validation errors
@app.exception_handler(ValidationError)
async def validation_exception_handler(request: Request, exc: ValidationError):
    """
    Handle Pydantic validation errors with detailed, user-friendly messages.
    """
    errors = []
    for error in exc.errors():
        field_path = " -> ".join(str(loc) for loc in error["loc"])
        errors.append({
            "field": field_path,
            "message": error["msg"],
            "type": error["type"],
            "input": str(error.get("input", ""))[:100]  # Truncate long inputs
        })
    
    return JSONResponse(
        status_code=422,
        content={
            "detail": "Validation error",
            "errors": errors
        }
    )

# Global exception handler for ValueError (from model validators)
@app.exception_handler(ValueError)
async def value_error_handler(request: Request, exc: ValueError):
    """
    Handle ValueError exceptions from model validators.
    """
    return JSONResponse(
        status_code=422,
        content={
            "detail": "Validation error",
            "message": str(exc)
        }
    )

# Include routers
app.include_router(trigger.router, prefix="/api/v1", tags=["Trigger"])
app.include_router(pipeline.router, prefix="/api/v1", tags=["Pipelines"])
app.include_router(handler.router, prefix="/api/v1", tags=["Handlers"])
app.include_router(worker.router, prefix="/api/v1", tags=["Worker"])
app.include_router(nodes.router, prefix="/api/v1", tags=["Nodes"])
app.include_router(websocket.router, prefix="/api/v1", tags=["WebSocket"])

# Health check endpoint
@app.get("/health", tags=["Health"])
async def health_check():
    return {"status": "healthy"}
