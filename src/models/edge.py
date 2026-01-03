#!/usr/bin/env python3
"""
Edge model for defining connections between nodes in a pipeline.
"""
import uuid
from typing import Optional
from pydantic import BaseModel, Field, model_validator, ConfigDict


class Edge(BaseModel):
    """
    Represents a connection between two nodes in a pipeline.

    Supports multi-port connections by specifying source and target ports.
    """

    model_config = ConfigDict(
        validate_assignment=True,
        extra="forbid",
        json_schema_extra={
            "example": {
                "source": "550e8400-e29b-41d4-a716-446655440000",
                "target": "550e8400-e29b-41d4-a716-446655440001",
                "source_port": "http_out",
                "target_port": "http_in",
            }
        },
    )

    source: uuid.UUID = Field(..., description="ID of the source node")
    target: uuid.UUID = Field(..., description="ID of the target node")
    source_port: Optional[str] = Field(
        default=None,
        max_length=50,
        pattern=r"^[a-z]+_out$",
        description="ID of the source port (format: {type}_out, e.g., 'http_out')",
    )
    target_port: Optional[str] = Field(
        default=None,
        max_length=50,
        pattern=r"^[a-z]+_in$",
        description="ID of the target port (format: {type}_in, e.g., 'http_in')",
    )

    @model_validator(mode="after")
    def validate_edge(self) -> "Edge":
        """Validate edge consistency."""
        # Source and target cannot be the same (self-loop)
        if self.source == self.target:
            raise ValueError("Edge cannot connect a node to itself (self-loop)")

        # If one port is specified, both should ideally be specified for clarity
        # But we allow partial specification for backwards compatibility

        # Validate port type consistency if both are specified
        if self.source_port and self.target_port:
            source_type = self.source_port.replace("_out", "")
            target_type = self.target_port.replace("_in", "")
            if source_type != target_type:
                raise ValueError(
                    f"Port type mismatch: source port '{self.source_port}' "
                    f"cannot connect to target port '{self.target_port}'"
                )

        return self
