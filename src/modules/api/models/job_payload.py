#!/usr/bin/env python3
from typing import Optional, List, Dict
from pydantic import BaseModel

class JobPayload(BaseModel):
    http: Optional[List[Dict]] = None
    ip: Optional[List[str]] = None