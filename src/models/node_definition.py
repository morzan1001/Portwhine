#!/usr/bin/env python3
"""
Node Definition models for providing rich metadata about nodes (workers and triggers).
This enables the frontend to dynamically render nodes without hardcoded information.
"""
from typing import List, Optional, Any
from pydantic import BaseModel, Field

from pydantic import ConfigDict
from models.types import InputOutputType, WorkerCategory, FieldType, NodeType


class FieldDefinition(BaseModel):
    """Definition of a configuration field for a node."""

    name: str = Field(..., description="Field name/key (matches the Pydantic field name)")
    label: str = Field(..., description="Human-readable label for UI")
    type: FieldType = Field(..., description="Field data type for UI rendering")
    description: str = Field(default="", description="Help text for the field")
    required: bool = Field(default=False, description="Whether the field is required")
    default: Optional[Any] = Field(default=None, description="Default value")
    options: List[str] = Field(default_factory=list, description="Options for select/multiselect types")
    placeholder: Optional[str] = Field(default=None, description="Placeholder text for input")
    validation_pattern: Optional[str] = Field(default=None, description="Regex pattern for validation")
    min_value: Optional[float] = Field(default=None, description="Minimum value for numeric fields")
    max_value: Optional[float] = Field(default=None, description="Maximum value for numeric fields")


class PortDefinition(BaseModel):
    """Definition of an input or output port on a node."""

    id: str = Field(..., description="Unique port identifier (e.g., 'ip_in', 'http_out')")
    label: str = Field(..., description="Human-readable label for UI")
    data_type: InputOutputType = Field(..., description="Type of data this port handles")
    description: str = Field(default="", description="Help text for the port")
    required: bool = Field(default=True, description="Whether this port must be connected")
    multiple: bool = Field(default=False, description="Whether multiple connections are allowed")


class NodeDefinition(BaseModel):
    """
    Complete definition of a node type.

    This is the single source of truth for node metadata, used by:
    - Frontend for dynamic rendering and configuration
    - Backend for validation and documentation
    - API documentation
    """

    # Identity
    id: str = Field(..., description="Unique node type identifier (class name, e.g., 'NmapWorker')")
    name: str = Field(..., description="Human-readable name (e.g., 'Nmap Scanner')")
    description: str = Field(..., description="Detailed description of what this node does")

    # Classification
    node_type: NodeType = Field(..., description="Whether this is a trigger or worker node")
    category: Optional[WorkerCategory] = Field(default=None, description="Worker category (None for triggers)")

    # UI Properties
    icon: str = Field(default="default", description="Icon identifier for the UI")
    color: str = Field(default="#6366F1", description="Primary color for the node (hex)")

    # Ports (inputs/outputs)
    inputs: List[PortDefinition] = Field(default_factory=list, description="Input ports")
    outputs: List[PortDefinition] = Field(default_factory=list, description="Output ports")

    # Configuration
    config_fields: List[FieldDefinition] = Field(default_factory=list, description="Configurable fields")

    # Runtime
    image_name: str = Field(..., description="Docker image name (e.g., 'nmap:1.0')")
    supports_multiple_instances: bool = Field(
        default=True, description="Whether multiple instances can run in parallel"
    )
    max_instances: Optional[int] = Field(
        default=None, description="Maximum number of parallel instances (None = unlimited)"
    )

    model_config = ConfigDict(use_enum_values=True)
