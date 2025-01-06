#!/usr/bin/env python3
from typing import Optional, List, Union, Dict, Any
from pydantic import BaseModel, IPvAnyAddress, IPvAnyNetwork, model_serializer

class JobPayload(BaseModel):
    http: Optional[List[Dict]] = None
    ip: Optional[List[Union[IPvAnyAddress, IPvAnyNetwork]]] = None