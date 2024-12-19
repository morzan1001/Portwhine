from pydantic import BaseModel
from typing import List, Optional

class WorkerConfig(BaseModel):
    worker_name: str
    children: Optional[List['WorkerConfig']] = []
