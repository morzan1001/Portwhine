#!/usr/bin/env python3
import uuid
from pydantic import BaseModel, PrivateAttr, model_serializer, IPvAnyAddress
from typing import Any, List, ClassVar
from api.models.types import InputOutputType, NodeStatus

class TriggerConfig(BaseModel):
    _id: uuid.UUID = PrivateAttr(default_factory=uuid.uuid4)
    _status: str = PrivateAttr(default=NodeStatus.STOPPED)

    def __init__(self, **data: Any):
        super().__init__(**data)
        if 'id' in data:
            self._id = uuid.UUID(data['id'])
        if 'status' in data: 
            self._status = NodeStatus(data['status'])
    
    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        data = {self.__class__.__name__: self.__dict__}
        data[self.__class__.__name__]['id'] = str(self._id)
        data[self.__class__.__name__]['status'] = self._status
        data[self.__class__.__name__]['output'] = self.__class__.output
        return data

class IPAddressTrigger(TriggerConfig):
    """
    Trigger that accepts a list of IP addresses.
    """
    ip_addresses: List[IPvAnyAddress]
    image_name: ClassVar[str] = ""
    output: ClassVar[List[InputOutputType]] = [InputOutputType.IP]

    def model_dump(self, *args, **kwargs):
        data = super().model_dump(*args, **kwargs)
        data['ip_addresses'] = [str(ip) for ip in self.ip_addresses]
        return data

class CertstreamTrigger(TriggerConfig):
    """
    Trigger that accepts a regex pattern.
    """
    regex: str
    image_name: ClassVar[str] = "certstream:1.0"
    output: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]