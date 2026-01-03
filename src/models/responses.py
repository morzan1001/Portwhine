#!/usr/bin/env python3
"""
API Response models for OpenAPI schema generation.

These models are used specifically for response_model in FastAPI routes
to generate accurate OpenAPI schemas. They mirror the serialized output
of the main models but with explicit field definitions.

Design:
- Response models include runtime fields (status, inputs, outputs, instanceHealth)
- Storage models (used internally) exclude runtime fields
- This separation ensures the API contract is clear and stable
"""
from typing import Any, Dict, List, Optional
from datetime import datetime
from pydantic import BaseModel, Field

from models.types import NodeStatus
from models.edge import Edge
from models.grid_position import GridPosition


# ============================================================================
# Response models for nested objects
# ============================================================================

class InstanceHealthResponse(BaseModel):
    """Response model for instance health - explicit fields for OpenAPI."""
    instance_number: int = Field(..., description="Instance number (0-indexed)")
    container_id: Optional[str] = Field(default=None, description="Docker container ID")
    container_name: Optional[str] = Field(default=None, description="Docker container name")
    status: str = Field(default="Unknown", description="Current status of the instance")
    started_at: Optional[str] = Field(default=None, description="When the instance started (ISO format)")
    finished_at: Optional[str] = Field(default=None, description="When the instance finished (ISO format)")
    exit_code: Optional[int] = Field(default=None, description="Exit code if the instance has finished")
    error_message: Optional[str] = Field(default=None, description="Error message if status is ERROR")
    jobs_processed: int = Field(default=0, description="Number of jobs processed by this instance")


# ============================================================================
# Response models for trigger and worker configuration
# These include RUNTIME fields (status, inputs, outputs) that are computed
# ============================================================================

class TriggerConfigResponse(BaseModel):
    """
    Response model for trigger configuration in API responses.
    
    Includes runtime fields that are computed from class variables:
    - status: Current runtime status
    - outputs: Output data types (from ClassVar)
    """
    id: str = Field(..., description="Unique trigger instance ID (UUID)")
    status: str = Field(default="Paused", description="Current runtime status of the trigger")
    outputs: List[str] = Field(default_factory=list, description="Output data types this trigger produces")
    gridPosition: GridPosition = Field(
        default_factory=GridPosition,
        description="Position on the canvas"
    )
    # Additional dynamic fields are allowed for trigger-specific config
    model_config = {"extra": "allow"}


class WorkerConfigResponse(BaseModel):
    """
    Response model for worker configuration in API responses.
    
    Includes runtime fields that are computed from class variables:
    - status: Current runtime status
    - inputs: Input data types (from ClassVar)
    - outputs: Output data types (from ClassVar)
    - instanceHealth: Health status per instance
    """
    id: str = Field(..., description="Unique worker instance ID (UUID)")
    status: str = Field(default="Paused", description="Current runtime status of the worker")
    inputs: List[str] = Field(default_factory=list, description="Input data types this worker accepts")
    outputs: List[str] = Field(default_factory=list, description="Output data types this worker produces")
    gridPosition: GridPosition = Field(
        default_factory=GridPosition,
        description="Position on the canvas"
    )
    numberOfInstances: int = Field(
        default=0,
        ge=0,
        le=100,
        description="Number of parallel instances"
    )
    instanceHealth: Optional[List[InstanceHealthResponse]] = Field(
        default=None,
        description="Health status per instance"
    )
    # Additional dynamic fields are allowed for worker-specific config
    model_config = {"extra": "allow"}


# ============================================================================
# Storage/Patch models for input (what the frontend sends)
# These do NOT include runtime fields - they are stripped/ignored
# ============================================================================

class TriggerConfigInput(BaseModel):
    """
    Input model for trigger configuration (what frontend sends).
    
    Does NOT include runtime fields - they will be ignored if sent.
    """
    id: Optional[str] = Field(default=None, description="Trigger instance ID (UUID or 'new-xxx' for new triggers)")
    gridPosition: Optional[GridPosition] = Field(default=None, description="Position on the canvas")
    # Additional dynamic fields are allowed for trigger-specific config
    model_config = {"extra": "allow"}


class WorkerConfigInput(BaseModel):
    """
    Input model for worker configuration (what frontend sends).
    
    Does NOT include runtime fields - they will be ignored if sent.
    """
    id: Optional[str] = Field(default=None, description="Worker instance ID (UUID or 'new-xxx' for new workers)")
    gridPosition: Optional[GridPosition] = Field(default=None, description="Position on the canvas")
    numberOfInstances: int = Field(default=0, ge=0, le=100, description="Number of parallel instances")
    # Additional dynamic fields are allowed for worker-specific config
    model_config = {"extra": "allow"}


class PipelineResponse(BaseModel):
    """
    Response model for pipeline data (what the API returns).
    
    This is the typed response that matches what the Pipeline.ser_model() produces.
    Used as response_model in FastAPI to generate accurate OpenAPI schemas.
    
    Includes runtime fields in trigger/worker that are computed from ClassVars.
    """
    id: str = Field(..., description="Unique pipeline ID (UUID)")
    status: str = Field(
        default="Stopped",
        description="Current pipeline status"
    )
    name: str = Field(..., min_length=1, max_length=100, description="Pipeline name")
    trigger: Optional[Dict[str, TriggerConfigResponse]] = Field(
        default=None,
        description="The trigger that initiates this pipeline (keyed by trigger class name)"
    )
    worker: List[Dict[str, WorkerConfigResponse]] = Field(
        default_factory=list,
        description="List of workers in this pipeline (keyed by worker class name)"
    )
    edges: List[Edge] = Field(
        default_factory=list,
        description="Connections between nodes"
    )


class PipelineListItem(BaseModel):
    """Response model for pipeline list items."""
    id: str = Field(..., description="Pipeline ID")
    name: str = Field(..., description="Pipeline name")
    status: str = Field(default="Unknown", description="Pipeline status")


class MessageResponse(BaseModel):
    """Generic message response."""
    detail: str = Field(..., description="Response message")


class DeleteResponse(BaseModel):
    """Response for delete operations."""
    detail: str = Field(default="Pipeline deleted successfully", description="Deletion status message")


# ============================================================================
# Response models for trigger/worker config endpoints
# ============================================================================

class NodeConfigExampleResponse(BaseModel):
    """Response model for trigger/worker configuration example."""
    description: str = Field(..., description="Description of the node")
    example: Dict[str, Any] = Field(..., description="Example configuration")


class NodeNotFoundResponse(BaseModel):
    """Response model when a node is not found."""
    error: str = Field(..., description="Error message")
