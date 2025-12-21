#!/usr/bin/env python3
from typing import List, Optional, Union
from pydantic import BaseModel, IPvAnyAddress, IPvAnyNetwork, Field

class HttpTarget(BaseModel):
    url: str
    method: str = "GET"
    headers: dict = {}

class IpTarget(BaseModel):
    ip: Union[IPvAnyAddress, IPvAnyNetwork]
    port: Optional[int] = None

class JobPayload(BaseModel):
    http: List[HttpTarget] = Field(default_factory=list)
    ip: List[IpTarget] = Field(default_factory=list)