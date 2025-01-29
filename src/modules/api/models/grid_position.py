#!/usr/bin/env python3
from pydantic import BaseModel, Field, model_serializer


class GridPosition(BaseModel):
    x: float = Field(default=0.0)
    y: float = Field(default=0.0)

    @model_serializer()
    def ser_model(self) -> dict[str, float]:
        return {'x': self.x, 'y': self.y}