#!/usr/bin/env python3
import uuid
from typing import Dict, Any, Optional
from pydantic import BaseModel
from models.types import NodeStatus
from models.job_payload import JobPayload

class WorkerResult(BaseModel):
    run_id: uuid.UUID
    pipeline_id: uuid.UUID
    node_id: uuid.UUID
    status: NodeStatus
    output_payload: Optional[JobPayload] = None
    raw_data: Optional[Dict[str, Any]] = None
    error: Optional[str] = None
