#!/usr/bin/env python3
import logging
from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware

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

# FastAPI app with metadata
app = FastAPI(
    title=app_metadata["title"],
    description=app_metadata["description"],
    version=app_metadata["version"],
    license_info=app_metadata["license_info"],
    openapi_tags=tags_metadata
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.middleware("http")
async def set_secure_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["Content-Security-Policy"] = "default-src 'self'; script-src 'self' 'unsafe-inline' cdn.jsdelivr.net; style-src 'self' cdn.jsdelivr.net; img-src 'self' fastapi.tiangolo.com; upgrade-insecure-requests;"
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
    response.headers["Permissions-Policy"] = "geolocation=(), microphone=(), camera=()"
    response.headers["Strict-Transport-Security"] = "max-age=31536000; includeSubDomains, preload"
    return response

# Include routers
app.include_router(trigger.router, prefix="/api/v1", tags=["Trigger"])
app.include_router(pipeline.router, prefix="/api/v1", tags=["Pipelines"])
app.include_router(handler.router, prefix="/api/v1", tags=["Handlers"])
app.include_router(worker.router, prefix="/api/v1", tags=["Worker"])

queue_thread = start_queue_thread()
container_health_thread = start_container_health_thread()

# Health check endpoint
@app.get("/health", tags=["Health"])
async def health_check():
    return {"status": "healthy"}