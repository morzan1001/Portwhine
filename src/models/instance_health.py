#!/usr/bin/env python3
from pydantic import BaseModel, Field, model_serializer

from models.types import NodeStatus

class InstanceHealth(BaseModel):
    number: int = Field(default=0)
    health: NodeStatus

    @model_serializer()
    def ser_model(self) -> dict[str, float]:
        return {'number': self.number, 'health': self.health}