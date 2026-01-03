#!/usr/bin/env python3
import uuid
from datetime import datetime, timezone
from typing import Dict, Optional
from pydantic import BaseModel, Field
from models.types import NodeStatus


class NodeRunState(BaseModel):
    status: NodeStatus = NodeStatus.PENDING
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    error: Optional[str] = None


class PipelineRun(BaseModel):
    id: uuid.UUID = Field(default_factory=uuid.uuid4)
    pipeline_id: uuid.UUID
    start_time: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    end_time: Optional[datetime] = None
    status: NodeStatus = NodeStatus.RUNNING
    node_states: Dict[str, NodeRunState] = Field(default_factory=dict)
