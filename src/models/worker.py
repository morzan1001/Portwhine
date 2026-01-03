#!/usr/bin/env python3
"""
Worker models for pipeline processing nodes.

Workers are nodes that process input data and produce output data.
Each worker type defines its input/output types, Docker image, and configuration fields.
"""
import uuid
from typing import Any, ClassVar, Dict, List, Optional

from pydantic import BaseModel, Field, PrivateAttr, model_serializer, model_validator, field_validator, ConfigDict

from models.types import InputOutputType, NodeStatus, WorkerCategory
from models.grid_position import GridPosition
from models.instance_health import InstanceHealth


class WorkerConfig(BaseModel):
    """
    Base configuration for all worker nodes.

    Workers process input data from connected nodes and produce output data
    that can be consumed by downstream nodes.

    Class Variables (defined on subclasses):
        inputs: List of input data types this worker accepts
        outputs: List of output data types this worker produces
        image_name: Docker image name for the worker container
        category: Node category for UI organization
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
    numberOfInstances: int = Field(
        default=0, ge=0, le=100, description="Number of parallel instances"  # Reasonable upper limit
    )
    instanceHealth: Optional[List[InstanceHealth]] = Field(default=None, description="Health status per instance")

    # Class variables for node metadata (overridden in subclasses)
    inputs: ClassVar[List[InputOutputType]] = []
    outputs: ClassVar[List[InputOutputType]] = []
    image_name: ClassVar[str] = ""
    category: ClassVar[WorkerCategory] = WorkerCategory.UTILITY
    display_name: ClassVar[str] = "Worker"

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
        """Unique identifier for this worker instance."""
        return self._id

    @property
    def status(self) -> NodeStatus:
        """Current runtime status of this worker."""
        return self._status

    @status.setter
    def status(self, value: NodeStatus) -> None:
        """Set the runtime status of this worker."""
        self._status = value

    @classmethod
    def get_config_fields(cls) -> Dict[str, Any]:
        """
        Extract configuration field definitions from the Pydantic model.

        Returns a dict of field name -> field info for all configurable fields
        (excludes base class fields like gridPosition, numberOfInstances).
        """
        base_fields = {"gridPosition", "numberOfInstances", "instanceHealth"}
        return {name: field for name, field in cls.model_fields.items() if name not in base_fields}

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        """Serialize the worker for API responses (includes runtime data)."""
        data = self.ser_for_storage()
        # Unwrap from {ClassName: data} format
        inner = data[self.__class__.__name__]
        # Add runtime/computed fields for API responses
        inner["status"] = self._status.value if isinstance(self._status, NodeStatus) else self._status
        inner["inputs"] = [t.value for t in self.__class__.inputs]
        inner["outputs"] = [t.value for t in self.__class__.outputs]
        inner["instanceHealth"] = [h.model_dump() for h in self.instanceHealth] if self.instanceHealth else None
        return {self.__class__.__name__: inner}

    def ser_for_storage(self) -> dict[str, Any]:
        """
        Serialize the worker for database storage.
        
        Only includes persistable data:
        - id, gridPosition, numberOfInstances
        - Worker-specific config fields
        
        Excludes runtime fields:
        - status, inputs, outputs, instanceHealth
        """
        data = {
            "id": str(self._id),
            "gridPosition": self.gridPosition.model_dump(),
            "numberOfInstances": self.numberOfInstances,
        }

        # Add configuration fields (worker-specific)
        for name in self.get_config_fields():
            if hasattr(self, name):
                data[name] = getattr(self, name)

        return {self.__class__.__name__: data}

    @model_validator(mode="after")
    def validate_worker_hierarchy(cls, values):
        return values


class FFUFWorker(WorkerConfig):
    """
    Worker that performs directory and file fuzzing on HTTP endpoints.

    Uses FFUF (Fuzz Faster U Fool) to discover hidden paths, files, and endpoints
    on web applications by testing with wordlists.
    """

    inputs: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    image_name: ClassVar[str] = "ffuf:1.0"
    category: ClassVar[WorkerCategory] = WorkerCategory.SCANNER
    display_name: ClassVar[str] = "FFUF Fuzzer"

    wordlist: str = Field(
        default="/usr/share/wordlists/common.txt",
        min_length=1,
        max_length=500,
        description="Path to the wordlist file inside the container",
    )
    extensions: Optional[str] = Field(
        default=None,
        max_length=200,
        description="Comma-separated list of file extensions to fuzz (e.g., 'php,html,js')",
    )
    recursive: bool = Field(default=False, description="Enable recursive fuzzing of discovered directories")

    @field_validator("wordlist")
    @classmethod
    def validate_wordlist(cls, v: str) -> str:
        from models.validators import validate_wordlist_path

        return validate_wordlist_path(v)

    @field_validator("extensions")
    @classmethod
    def validate_extensions(cls, v: Optional[str]) -> Optional[str]:
        from models.validators import validate_file_extensions

        return validate_file_extensions(v)


class HumbleWorker(WorkerConfig):
    """
    Worker that analyzes HTTP security headers.

    Uses Humble to check for missing or misconfigured security headers
    like CSP, HSTS, X-Frame-Options, etc.
    """

    inputs: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    image_name: ClassVar[str] = "humble:1.0"
    category: ClassVar[WorkerCategory] = WorkerCategory.ANALYZER
    display_name: ClassVar[str] = "Humble Header Analyzer"


class ScreenshotWorker(WorkerConfig):
    """
    Worker that captures screenshots of web pages.

    Uses a headless browser to render and capture screenshots of HTTP endpoints
    for visual analysis and documentation.
    """

    inputs: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    image_name: ClassVar[str] = "screenshot:1.0"
    category: ClassVar[WorkerCategory] = WorkerCategory.UTILITY
    display_name: ClassVar[str] = "Screenshot"

    resolution: str = Field(
        default="1920x1080",
        pattern=r"^\d{3,5}x\d{3,5}$",
        description="Screen resolution for the screenshot (WIDTHxHEIGHT)",
    )

    @field_validator("resolution")
    @classmethod
    def validate_resolution(cls, v: str) -> str:
        from models.validators import validate_resolution

        return validate_resolution(v)


class TestSSLWorker(WorkerConfig):
    """
    Worker that tests SSL/TLS configurations.

    Uses testssl.sh to analyze SSL/TLS settings, certificate validity,
    cipher suites, and known vulnerabilities.
    """

    inputs: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    image_name: ClassVar[str] = "testssl:1.0"
    category: ClassVar[WorkerCategory] = WorkerCategory.ANALYZER
    display_name: ClassVar[str] = "TestSSL Analyzer"


class WebAppAnalyzerWorker(WorkerConfig):
    """
    Worker that detects web technologies and frameworks.

    Uses Wappalyzer to identify CMS, frameworks, programming languages,
    JavaScript libraries, and other technologies used by web applications.
    """

    inputs: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    image_name: ClassVar[str] = "webappanalyzer:1.0"
    category: ClassVar[WorkerCategory] = WorkerCategory.ANALYZER
    display_name: ClassVar[str] = "Web App Analyzer"


class NmapWorker(WorkerConfig):
    """
    Worker that performs network port scanning and service detection.

    Uses Nmap to discover open ports, running services, and OS information
    on target IP addresses. Can output both IP and HTTP targets.
    """

    inputs: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.IP, InputOutputType.HTTP]
    image_name: ClassVar[str] = "nmap:1.0"
    category: ClassVar[WorkerCategory] = WorkerCategory.SCANNER
    display_name: ClassVar[str] = "Nmap Scanner"

    ports: str = Field(
        default="-p-",
        max_length=100,
        description="Port specification (e.g., '-p-' for all ports, '-p1-1000' for range)",
    )
    arguments: str = Field(
        default="-A", max_length=500, description="Additional Nmap arguments (e.g., '-A' for aggressive scan)"
    )
    custom_command: Optional[str] = Field(
        default=None,
        max_length=1000,
        description="Full custom Nmap command. Use {{target}} as placeholder. Must output XML to stdout (-oX -).",
    )

    @field_validator("ports")
    @classmethod
    def validate_ports(cls, v: str) -> str:
        from models.validators import validate_port_range

        return validate_port_range(v)

    @field_validator("arguments")
    @classmethod
    def validate_arguments(cls, v: str) -> str:
        from models.validators import validate_nmap_arguments

        return validate_nmap_arguments(v)

    @field_validator("custom_command")
    @classmethod
    def validate_custom_cmd(cls, v: Optional[str]) -> Optional[str]:
        from models.validators import validate_custom_nmap_command

        return validate_custom_nmap_command(v)

    @model_validator(mode="after")
    def validate_nmap_config(self) -> "NmapWorker":
        """Ensure custom_command is used exclusively or ports/arguments are used."""
        if self.custom_command:
            # When custom_command is set, ports and arguments are ignored
            # but we don't need to error, just note that they're overridden
            pass
        return self


class ResolverWorker(WorkerConfig):
    """
    Worker that resolves domain names to IP addresses.

    Performs DNS resolution to convert HTTP URLs or domain names
    into IP addresses for further scanning.
    """

    inputs: ClassVar[List[InputOutputType]] = [InputOutputType.HTTP]
    outputs: ClassVar[List[InputOutputType]] = [InputOutputType.IP]
    image_name: ClassVar[str] = "resolver:1.0"
    category: ClassVar[WorkerCategory] = WorkerCategory.UTILITY
    display_name: ClassVar[str] = "DNS Resolver"

    use_internal: bool = Field(default=False, description="Use internal DNS resolver instead of system default")
