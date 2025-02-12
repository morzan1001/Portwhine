#!/usr/bin/env python3
import uuid
from pydantic import BaseModel, PrivateAttr, Field, model_serializer, field_validator, IPvAnyAddress, IPvAnyNetwork
from typing import Any, List, ClassVar, Optional, Union
from api.models.types import InputOutputType, NodeStatus
from api.models.grid_position import GridPosition


class TriggerConfig(BaseModel):
    _id: uuid.UUID = PrivateAttr(default_factory=uuid.uuid4)
    _status: str = PrivateAttr(default=NodeStatus.PAUSED)
    gridPosition: GridPosition = Field(default_factory=GridPosition)

    def __init__(self, **data: Any):
        super().__init__(**data)
        if 'id' in data:
            self._id = uuid.UUID(data['id'])
        if 'status' in data:
            self._status = NodeStatus(data['status'])

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        data = {self.__class__.__name__: self.__dict__}

        if isinstance(self.gridPosition, GridPosition):
            data[self.__class__.__name__]['gridPosition'] = self.gridPosition.ser_model()

        data[self.__class__.__name__]['id'] = str(self._id)
        data[self.__class__.__name__]['status'] = self._status
        data[self.__class__.__name__]['output'] = self.__class__.output
        return data

class IPAddressTrigger(TriggerConfig):
    """
    Trigger that accepts a list of IP addresses, single IP address, a list of networks, or single network. The repetition (seconds) defines if the trigger should start a scan repetitively.
    """
    ip_addresses: List[Union[IPvAnyAddress, IPvAnyNetwork]] = Field(default_factory=list)
    image_name: ClassVar[str] = "ipaddress:1.0"
    output: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    repetition: Optional[int] = None

    @field_validator('ip_addresses', mode="before")
    @classmethod
    def ensure_list(cls, v) -> list:
        if not isinstance(v, list):
            return [v]
        return v

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        data = super().ser_model()
        if isinstance(self.ip_addresses, list):
            data[self.__class__.__name__]['ip_addresses'] = [str(ip) for ip in self.ip_addresses]
        else:
            data[self.__class__.__name__]['ip_addresses'] = str(self.ip_addresses)
        return data

class CertstreamTrigger(TriggerConfig):
    """
    Trigger that accepts a regex pattern. The trigger monitors certificate transparency logs for new certificates that match the regex pattern.
    """
    regex: str
    image_name: ClassVar[str] = "certstream:1.0"
    output: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]