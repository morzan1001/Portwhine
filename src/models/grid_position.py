#!/usr/bin/env python3
"""
Grid position model for node placement on the canvas.
"""
from pydantic import BaseModel, Field


class GridPosition(BaseModel):
    """
    Position of a node on the pipeline canvas.

    Uses floating point coordinates to support smooth dragging and zooming.
    """

    x: float = Field(default=0.0, description="Horizontal position on canvas")
    y: float = Field(default=0.0, description="Vertical position on canvas")

    def __eq__(self, other: object) -> bool:
        if not isinstance(other, GridPosition):
            return False
        return self.x == other.x and self.y == other.y

    def __hash__(self) -> int:
        return hash((self.x, self.y))
