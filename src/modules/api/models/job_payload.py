#!/usr/bin/env python3
from pydantic import BaseModel
from typing import Optional, List, Dict

class JobPayload(BaseModel):
    http: Optional[List[Dict]] = None
    ip: Optional[List[str]] = None