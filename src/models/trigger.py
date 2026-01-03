#!/usr/bin/env python3
"""
Trigger models for pipeline initiation nodes.

Triggers are nodes that start pipeline execution based on external events
or scheduled intervals. They have no inputs and produce output data.
"""
import uuid
from typing import Any, List, ClassVar, Dict, Optional, Union

from pydantic import (
    BaseModel,
    PrivateAttr,
    Field,
    model_serializer,
    field_validator,
    model_validator,
    IPvAnyAddress,
    IPvAnyNetwork,
    ConfigDict,
)

from models.types import InputOutputType, NodeStatus
from models.grid_position import GridPosition


class TriggerConfig(BaseModel):
    """
    Base configuration for all trigger nodes.

    Triggers initiate pipeline execution by producing output data.
    Unlike workers, triggers have no inputs - they generate data from
    external sources or configurations.

    Class Variables (defined on subclasses):
        outputs: List of output data types this trigger produces
        image_name: Docker image name for the trigger container
        display_name: Human-readable name for the UI
    """

    model_config = ConfigDict(
        validate_assignment=True,  # Validate on attribute assignment
        str_strip_whitespace=True,  # Strip whitespace from strings
        extra="forbid",  # Strict validation - reject unexpected fields
    )

    # Private attributes for runtime state
    _id: uuid.UUID = PrivateAttr(default_factory=uuid.uuid4)
    _status: NodeStatus = PrivateAttr(default=NodeStatus.PAUSED)

    # Serialized configuration
    gridPosition: GridPosition = Field(default_factory=GridPosition, description="Position on the canvas")

    # Class variables for node metadata (overridden in subclasses)
    outputs: ClassVar[List[InputOutputType]] = []
    image_name: ClassVar[str] = ""
    display_name: ClassVar[str] = "Trigger"

    def __init__(self, **data: Any):
        id_val = data.pop("id", None)
        status_val = data.pop("status", None)
        super().__init__(**data)
        if id_val:
            self._id = uuid.UUID(id_val)
        if status_val:
            self._status = NodeStatus(status_val)

    @property
    def id(self) -> uuid.UUID:
        """Unique identifier for this trigger instance."""
        return self._id

    @property
    def status(self) -> NodeStatus:
        """Current runtime status of this trigger."""
        return self._status

    @status.setter
    def status(self, value: NodeStatus) -> None:
        """Set the runtime status of this trigger."""
        self._status = value

    @classmethod
    def get_config_fields(cls) -> Dict[str, Any]:
        """
        Extract configuration field definitions from the Pydantic model.

        Returns a dict of field name -> field info for all configurable fields
        (excludes base class fields like gridPosition).
        """
        base_fields = {"gridPosition"}
        return {name: field for name, field in cls.model_fields.items() if name not in base_fields}

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        """Serialize the trigger for API responses (includes runtime data)."""
        data = self.ser_for_storage()
        # Unwrap from {ClassName: data} format
        inner = data[self.__class__.__name__]
        # Add runtime/computed fields for API responses
        inner["status"] = self._status.value if isinstance(self._status, NodeStatus) else self._status
        inner["outputs"] = [t.value for t in self.__class__.outputs]
        return {self.__class__.__name__: inner}

    def ser_for_storage(self) -> dict[str, Any]:
        """
        Serialize the trigger for database storage.
        
        Only includes persistable data:
        - id, gridPosition
        - Trigger-specific config fields
        
        Excludes runtime fields:
        - status, outputs
        """
        data = {
            "id": str(self._id),
            "gridPosition": self.gridPosition.model_dump(),
        }

        # Add configuration fields (trigger-specific)
        for name in self.get_config_fields():
            if hasattr(self, name):
                value = getattr(self, name)
                # Handle special types
                if isinstance(value, (IPvAnyAddress, IPvAnyNetwork)):
                    data[name] = str(value)
                elif isinstance(value, list) and value and isinstance(value[0], (IPvAnyAddress, IPvAnyNetwork)):
                    data[name] = [str(v) for v in value]
                else:
                    data[name] = value

        return {self.__class__.__name__: data}


class IPAddressTrigger(TriggerConfig):
    """
    Trigger that initiates pipeline execution with IP addresses.

    Accepts a list of IP addresses or networks and optionally repeats
    the trigger at specified intervals. Useful for scheduled scanning
    of known infrastructure.
    """

    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    image_name: ClassVar[str] = "ipaddress:1.0"
    display_name: ClassVar[str] = "IP Address Trigger"

    ip_addresses: List[Union[IPvAnyAddress, IPvAnyNetwork]] = Field(
        default_factory=list,
        min_length=1,  # At least one IP required
        description="List of IP addresses or CIDR networks to trigger",
    )
    repetition: Optional[int] = Field(
        default=None,
        ge=60,
        le=86400 * 30,  # Max 30 days
        description="Interval in seconds between repeated triggers (60s - 30 days, None = once)",
    )

    @field_validator("ip_addresses", mode="before")
    @classmethod
    def ensure_list(cls, v) -> list:
        """Ensure ip_addresses is always a list."""
        if not isinstance(v, list):
            return [v]
        return v

    @model_validator(mode="after")
    def validate_ip_addresses(self) -> "IPAddressTrigger":
        """Validate that IP addresses are provided."""
        if not self.ip_addresses:
            raise ValueError("At least one IP address or network is required")
        # Limit to prevent abuse
        if len(self.ip_addresses) > 10000:
            raise ValueError("Maximum 10,000 IP addresses/networks allowed")
        return self


class CertstreamTrigger(TriggerConfig):
    """
    Trigger that monitors Certificate Transparency logs in real-time.

    Watches for newly issued SSL/TLS certificates matching a regex pattern.
    Useful for discovering new domains and subdomains as they appear.
    """

    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    image_name: ClassVar[str] = "certstream:1.0"
    display_name: ClassVar[str] = "Certstream Trigger"

    regex: str = Field(
        ...,
        min_length=1,
        max_length=1000,
        description="Regex pattern to match against certificate common names and SANs",
    )

    @field_validator("regex")
    @classmethod
    def validate_regex(cls, v: str) -> str:
        """Validate that the regex pattern is valid."""
        from models.validators import validate_regex_pattern

        return validate_regex_pattern(v, "regex")
