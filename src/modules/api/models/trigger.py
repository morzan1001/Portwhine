#!/usr/bin/env python3
import uuid
from pydantic import BaseModel, Field, model_serializer, IPvAnyAddress
from typing import Any, List, ClassVar
from api.models.types import InputOutputType

class TriggerConfig(BaseModel):
    id: uuid.UUID = Field(default_factory=uuid.uuid4)
    
    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        data = {self.__class__.__name__: self.__dict__}
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