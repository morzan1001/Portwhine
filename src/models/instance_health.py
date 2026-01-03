#!/usr/bin/env python3
"""
Instance health model for tracking the status of individual node instances.
"""
from typing import Optional
from datetime import datetime
from pydantic import BaseModel, Field, model_serializer

from models.types import NodeStatus


class InstanceHealth(BaseModel):
    """
    Health information for a single instance of a node (container).

    Each worker can have multiple instances running in parallel.
    This model tracks the health of each individual instance.
    """

    instance_number: int = Field(..., description="Instance number (0-indexed)")
    container_id: Optional[str] = Field(default=None, description="Docker container ID")
    container_name: Optional[str] = Field(default=None, description="Docker container name")
    status: NodeStatus = Field(default=NodeStatus.UNKNOWN, description="Current status of the instance")
    started_at: Optional[datetime] = Field(default=None, description="When the instance started")
    finished_at: Optional[datetime] = Field(default=None, description="When the instance finished (if applicable)")
    exit_code: Optional[int] = Field(default=None, description="Exit code if the instance has finished")
    error_message: Optional[str] = Field(default=None, description="Error message if status is ERROR")
    jobs_processed: int = Field(default=0, description="Number of jobs processed by this instance")

    @model_serializer()
    def ser_model(self) -> dict:
        return {
            "instance_number": self.instance_number,
            "container_id": self.container_id,
            "container_name": self.container_name,
            "status": self.status.value,
            "started_at": self.started_at.isoformat() if self.started_at else None,
            "finished_at": self.finished_at.isoformat() if self.finished_at else None,
            "exit_code": self.exit_code,
            "error_message": self.error_message,
            "jobs_processed": self.jobs_processed,
        }
