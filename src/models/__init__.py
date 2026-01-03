#!/usr/bin/env python3
"""
Portwhine Models Package.

This package contains all data models for the Portwhine pipeline system.
"""

# Core types and enums
from models.types import (
    InputOutputType,
    NodeStatus,
    WorkerCategory,
    FieldType,
    NodeType,
)

# Base models
from models.grid_position import GridPosition
from models.instance_health import InstanceHealth
from models.edge import Edge

# Pipeline models
from models.pipeline import Pipeline
from models.pipeline_run import PipelineRun
from models.job_payload import JobPayload
from models.worker_result import WorkerResult

# Worker models
from models.worker import (
    WorkerConfig,
    FFUFWorker,
    HumbleWorker,
    ScreenshotWorker,
    TestSSLWorker,
    WebAppAnalyzerWorker,
    NmapWorker,
    ResolverWorker,
)

# Trigger models
from models.trigger import (
    TriggerConfig,
    IPAddressTrigger,
    CertstreamTrigger,
)

# Node definition models (for API/UI)
from models.node_definition import (
    NodeDefinition,
    PortDefinition,
    FieldDefinition,
)

# Response models (for OpenAPI schema)
from models.responses import (
    PipelineResponse,
    PipelineListItem,
    MessageResponse,
    DeleteResponse,
    TriggerConfigResponse,
    WorkerConfigResponse,
    InstanceHealthResponse,
    NodeConfigExampleResponse,
)

# Node registry
from models.node_registry import NodeRegistry

__all__ = [
    # Types
    "InputOutputType",
    "NodeStatus",
    "WorkerCategory",
    "FieldType",
    "NodeType",
    # Base models
    "GridPosition",
    "InstanceHealth",
    "Edge",
    # Pipeline
    "Pipeline",
    "PipelineRun",
    "JobPayload",
    "WorkerResult",
    # Workers
    "WorkerConfig",
    "FFUFWorker",
    "HumbleWorker",
    "ScreenshotWorker",
    "TestSSLWorker",
    "WebAppAnalyzerWorker",
    "NmapWorker",
    "ResolverWorker",
    # Triggers
    "TriggerConfig",
    "IPAddressTrigger",
    "CertstreamTrigger",
    # Node definitions
    "NodeDefinition",
    "PortDefinition",
    "FieldDefinition",
    "NodeRegistry",
    # Response models
    "PipelineResponse",
    "PipelineListItem",
    "MessageResponse",
    "DeleteResponse",
    "TriggerConfigResponse",
    "WorkerConfigResponse",
    "InstanceHealthResponse",
    "NodeConfigExampleResponse",
]
