#!/usr/bin/env python3
import uuid
from pydantic import BaseModel

class Edge(BaseModel):
    source: uuid.UUID
    target: uuid.UUID
