#!/usr/bin/env python3
from pydantic import BaseModel, Field

class GridPosition(BaseModel):
    x: float = Field(default=0.0)
    y: float = Field(default=0.0)