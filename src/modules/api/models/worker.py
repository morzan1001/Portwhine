#!/usr/bin/env python3
import uuid
from pydantic import BaseModel, Field, PrivateAttr, model_serializer, model_validator
from typing import Any, ClassVar, List, Optional, Set
from api.models.types import InputOutputType, NodeStatus
from api.models.grid_position import GridPosition
from api.models.instance_health import InstanceHealth

class WorkerConfig(BaseModel):
    _id: uuid.UUID = PrivateAttr(default_factory=uuid.uuid4)
    _status: str = PrivateAttr(default=NodeStatus.PAUSED)
    children: Optional[List['WorkerConfig']] = None
    gridPosition: GridPosition = Field(default_factory=GridPosition)
    numberOfInstances: int = 0
    instanceHealth: Optional[List[InstanceHealth]] = None

    def __init__(self, **data: Any):
        super().__init__(**data)
        if 'id' in data:
            self._id = uuid.UUID(data['id'])
        if 'status' in data:
            self._status = NodeStatus(data['status'])

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        data = {
            'id': str(self._id),
            'status': self._status,
            'input': self.__class__.input,
            'output': self.__class__.output,
            'children': [child.ser_model() for child in self.children] if self.children else None,
            'gridPosition': self.gridPosition.ser_model(),
            'instanceHealth': self.instanceHealth.ser_model() if self.instanceHealth else None,
        }

        # Add all other attributes that are not serialized by default
        for key, value in self.__dict__.items():
            if key not in data:
                data[key] = value
        return {self.__class__.__name__: data}


    @model_validator(mode="after")
    def validate_worker_hierarchy(cls, values):
        children = values.children
        if children:
            parent_output_set: Set[InputOutputType] = set(values.output)
            for child in children:
                child_input_set: Set[InputOutputType] = set(child.input)
                if not parent_output_set & child_input_set:
                    raise ValueError(f"Worker {values._id} output does not match any child {child._id} input")
        return values

class FFUFWorker(WorkerConfig):
    """
    Worker that performs fuzzing on HTTP endpoints.
    """
    input: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    output: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    image_name: ClassVar[str] = "ffuf:1.0"

class HumbleWorker(WorkerConfig):
    """
    Worker that analyzes HTTP headers.
    """
    input: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    output: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    image_name: ClassVar[str] = "humble:1.0"

class ScreenshotWorker(WorkerConfig):
    """
    Worker that takes screenshots of HTTP endpoints.
    """
    input: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    output: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    image_name: ClassVar[str] = "screenshot:1.0"

class TestSSLWorker(WorkerConfig):
    """
    Worker that tests SSL configurations on IP addresses.
    """
    input: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    output: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    image_name: ClassVar[str] = "testssl:1.0"

class WebAppAnalyzerWorker(WorkerConfig):
    """
    Worker that analyzes web applications on HTTP endpoints.
    """
    input: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    output: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    image_name: ClassVar[str] = "webappanalyzer:1.0"

class NmapWorker(WorkerConfig):
    """
    Worker that performs network mapping on IP addresses.
    """
    input: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    output: ClassVar[List[InputOutputType]] = [InputOutputType.IP, InputOutputType.HTTP]
    image_name: ClassVar[str] = "nmap:1.0"

class ResolverWorker(WorkerConfig):
    """
    Worker that resolves domain names to IP addresses.
    """
    input: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    output: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    image_name: ClassVar[str] = "resolver:1.0"
    use_internal: bool = Field(default=False)