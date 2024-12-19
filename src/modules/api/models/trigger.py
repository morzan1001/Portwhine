from pydantic import BaseModel
from typing import List

class IPAddressTrigger(BaseModel):
    ip_addresses: List[str]
    worker: WorkerConfig

class CertstreamTrigger(BaseModel):
    regex: str
    worker: WorkerConfig