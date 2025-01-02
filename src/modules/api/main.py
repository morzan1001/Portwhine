#!/usr/bin/env python3
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware

from api.queue import start_queue_thread
from utils.logger import LoggingModule
from .routers import trigger, pipeline, handler, worker
from api.docs.main_docs import app_metadata, tags_metadata

# Logger initializing
logger = LoggingModule.get_logger()

# FastAPI app with metadata
app = FastAPI(
    title=app_metadata["title"],
    description=app_metadata["description"],
    version=app_metadata["version"],
    license_info=app_metadata["license_info"],
    openapi_tags=tags_metadata
)

@app.middleware("http")
async def set_secure_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["Content-Security-Policy"] = "default-src 'self'"
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
    response.headers["Permissions-Policy"] = "geolocation=(), microphone=(), camera=()"
    response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains, preload"
    response.headers["Access-Control-Allow-Origin"] = "*"
    response.headers["Access-Control-Allow-Credentials"] = "true"
    response.headers["Access-Control-Allow-Methods"] = "GET, POST, PUT, DELETE, OPTIONS"
    return response

# Include routers
app.include_router(trigger.router, prefix="/api/v1", tags=["Trigger"])
app.include_router(pipeline.router, prefix="/api/v1", tags=["Pipelines"])
app.include_router(handler.router, prefix="/api/v1", tags=["Handlers"])
app.include_router(worker.router, prefix="/api/v1", tags=["Worker"])

queue_thread = start_queue_thread()

# Health check endpoint
@app.get("/health", tags=["Health"])
async def health_check():
    return {"status": "healthy"}