#!/usr/bin/env python3
"""
Pipeline model for orchestrating triggers, workers, and their connections.
"""
import uuid
from typing import Any, Dict, List, Optional, Set

from pydantic import (
    BaseModel,
    PrivateAttr,
    SerializeAsAny,
    model_serializer,
    model_validator,
    field_validator,
    Field,
    ConfigDict,
)

from models.edge import Edge
from models.trigger import TriggerConfig, CertstreamTrigger, IPAddressTrigger
from models.worker import (
    ResolverWorker,
    WorkerConfig,
    FFUFWorker,
    HumbleWorker,
    ScreenshotWorker,
    TestSSLWorker,
    WebAppAnalyzerWorker,
    NmapWorker,
)
from models.types import InputOutputType, NodeStatus
from models.validators import validate_name, validate_pipeline_graph
from utils.logger import LoggingModule

logger = LoggingModule.get_logger()


# Registry of valid trigger and worker classes
TRIGGER_CLASSES = [CertstreamTrigger, IPAddressTrigger]
WORKER_CLASSES = [
    FFUFWorker,
    HumbleWorker,
    ScreenshotWorker,
    TestSSLWorker,
    WebAppAnalyzerWorker,
    NmapWorker,
    ResolverWorker,
]


class Pipeline(BaseModel):
    """
    A pipeline defines a workflow of triggers, workers, and their connections.

    Pipelines must follow these rules:
    - A pipeline with workers must have exactly one trigger
    - All workers must be reachable from the trigger
    - Edges must connect compatible port types
    - No cycles are allowed
    - No self-loops are allowed
    """

    model_config = ConfigDict(
        validate_assignment=True,
        str_strip_whitespace=True,
    )

    _id: uuid.UUID = PrivateAttr(default_factory=uuid.uuid4)
    _status: NodeStatus = PrivateAttr(default=NodeStatus.STOPPED)

    name: str = Field(..., min_length=1, max_length=100, description="Pipeline name")
    trigger: Optional[SerializeAsAny[TriggerConfig]] = Field(
        default=None, description="The trigger that initiates this pipeline"
    )
    worker: Optional[List[SerializeAsAny[WorkerConfig]]] = Field(
        default=None, description="List of workers in this pipeline"
    )
    edges: List[Edge] = Field(default_factory=list, description="Connections between nodes")

    @field_validator('edges', mode='before')
    @classmethod
    def edges_null_to_empty_list(cls, v):
        """Convert null edges to empty list."""
        if v is None:
            return []
        return v

    def __init__(self, **data: Any):
        super().__init__(**data)
        if "id" in data:
            self._id = uuid.UUID(data["id"])
        if "status" in data:
            self._status = NodeStatus(data["status"])

    @property
    def id(self) -> uuid.UUID:
        return self._id

    @property
    def status(self) -> NodeStatus:
        return self._status

    @status.setter
    def status(self, value: NodeStatus) -> None:
        self._status = value

    @field_validator("name")
    @classmethod
    def validate_pipeline_name(cls, v: str) -> str:
        """Validate the pipeline name."""
        return validate_name(v, "Pipeline name")

    @field_validator("edges")
    @classmethod
    def validate_edges_limit(cls, v: List[Edge]) -> List[Edge]:
        """Limit the number of edges to prevent abuse."""
        if len(v) > 1000:
            raise ValueError("Maximum 1000 edges allowed per pipeline")
        return v

    @field_validator("worker")
    @classmethod
    def validate_workers_limit(cls, v: Optional[List[WorkerConfig]]) -> Optional[List[WorkerConfig]]:
        """Limit the number of workers to prevent abuse."""
        if v and len(v) > 100:
            raise ValueError("Maximum 100 workers allowed per pipeline")
        return v

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        """Serialize the pipeline for API responses (includes runtime data)."""
        data = {
            "id": str(self._id),
            "status": self._status.value if isinstance(self._status, NodeStatus) else self._status,
            "name": self.name,
            "trigger": self.trigger.ser_model() if self.trigger else None,
            "worker": [workers_data.ser_model() for workers_data in self.worker] if self.worker else [],
            "edges": [edge.model_dump() for edge in self.edges],
        }
        return data

    def ser_for_storage(self) -> dict[str, Any]:
        """
        Serialize the pipeline for database storage.
        
        Only includes persistable data. Runtime fields (status, inputs, outputs, 
        instanceHealth) are excluded and will be computed when reading.
        """
        data = {
            "id": str(self._id),
            "status": self._status.value if isinstance(self._status, NodeStatus) else self._status,
            "name": self.name,
            "trigger": self.trigger.ser_for_storage() if self.trigger else None,
            "worker": [w.ser_for_storage() for w in self.worker] if self.worker else [],
            "edges": [edge.model_dump() for edge in self.edges],
        }
        return data

    @model_validator(mode="before")
    def validate_trigger(cls, values: Dict[str, Any]) -> Dict[str, Any]:
        """Parse and validate trigger data using the trigger class registry."""
        trigger_data = values.get("trigger")
        logger.debug("Validating trigger: %s", trigger_data)  # pylint: disable=logging-too-many-args

        # Treat empty, None, or gridPosition-only trigger as no trigger
        if trigger_data is None or trigger_data == {}:
            values["trigger"] = None
            return values
        
        # Check if trigger_data only contains gridPosition (no actual trigger type)
        if isinstance(trigger_data, dict):
            non_position_keys = [k for k in trigger_data.keys() if k != "gridPosition"]
            if not non_position_keys:
                # Only gridPosition, no trigger type - treat as no trigger
                values["trigger"] = None
                return values

        if isinstance(trigger_data, (CertstreamTrigger, IPAddressTrigger)):
            # Already a trigger instance
            return values

        # Find matching trigger class
        for trigger_cls in TRIGGER_CLASSES:
            key = trigger_cls.__name__
            if key in trigger_data:
                try:
                    values["trigger"] = trigger_cls(**trigger_data[key])
                except Exception as e:
                    raise ValueError(f"Invalid {key} configuration: {e}") from e
                return values

        available_triggers = [cls.__name__ for cls in TRIGGER_CLASSES]
        raise ValueError(f"Unknown trigger type. Available types: {available_triggers}")

    @model_validator(mode="before")
    def validate_workers(cls, values: Dict[str, Any]) -> Dict[str, Any]:
        """Parse and validate worker data using the worker class registry."""
        workers_data = values.get("worker")
        logger.debug("Validating workers: %s", workers_data)  # pylint: disable=logging-too-many-args

        if not workers_data:
            return values

        def validate_worker(worker_data: Dict[str, Any]) -> WorkerConfig:
            """Validate a single worker configuration."""
            if isinstance(worker_data, tuple(WORKER_CLASSES)):
                # Already a worker instance
                return worker_data

            for worker_cls in WORKER_CLASSES:
                key = worker_cls.__name__
                if key in worker_data:
                    try:
                        worker_dict = worker_data[key].copy()
                        return worker_cls(**worker_dict)
                    except Exception as e:
                        raise ValueError(f"Invalid {key} configuration: {e}") from e

            available_workers = [cls.__name__ for cls in WORKER_CLASSES]
            raise ValueError(f"Unknown worker type. Available types: {available_workers}")

        validated_workers = [validate_worker(worker_data) for worker_data in workers_data]
        values["worker"] = validated_workers
        return values

    @model_validator(mode="after")
    def validate_pipeline(cls, pipeline: "Pipeline"):
        """
        Comprehensive pipeline validation:
        - Structural integrity (trigger/worker relationships)
        - Edge connectivity (nodes exist, ports compatible)
        - Graph structure (no orphans, valid DAG for execution)
        """
        trigger_data = pipeline.trigger
        workers_data = pipeline.worker
        edges_data = pipeline.edges

        # Empty pipeline is valid (initial state)
        if not trigger_data and not workers_data:
            if edges_data:
                raise ValueError("Edges cannot exist without nodes")
            return pipeline

        # Workers require a trigger
        if workers_data and not trigger_data:
            raise ValueError("Workers cannot exist without a trigger")

        # Build node lookup
        nodes: Dict[uuid.UUID, Any] = {}
        if trigger_data:
            nodes[trigger_data._id] = trigger_data
        if workers_data:
            for worker in workers_data:
                if worker._id in nodes:
                    raise ValueError(f"Duplicate node ID: {worker._id}")
                nodes[worker._id] = worker

        # Validate edges
        if edges_data:
            for edge in edges_data:
                source_node = nodes.get(edge.source)
                target_node = nodes.get(edge.target)

                if not source_node:
                    raise ValueError(f"Edge source node {edge.source} not found in pipeline")
                if not target_node:
                    raise ValueError(f"Edge target node {edge.target} not found in pipeline")

                # Validate port type compatibility
                source_outputs: Set[InputOutputType] = set(source_node.outputs)
                target_inputs: Set[InputOutputType] = set(target_node.inputs)

                # Ports are required in current versions.
                if edge.source_port is None or edge.target_port is None:
                    raise ValueError(
                        "Edge ports are required (source_port/target_port). "
                        "Please update the frontend to the latest version."
                    )

                # Extract port type from port name (e.g., "domain_out" -> "domain")
                source_port_type = edge.source_port.replace("_out", "")
                target_port_type = edge.target_port.replace("_in", "")

                # Check if the port types are compatible
                try:
                    source_io_type = InputOutputType(source_port_type)
                    target_io_type = InputOutputType(target_port_type)
                except ValueError as exc:
                    raise ValueError(
                        f"Invalid port types in edge: {edge.source_port} -> {edge.target_port}"
                    ) from exc

                if source_io_type not in source_outputs:
                    raise ValueError(
                        f"Node {source_node._id} does not have output port '{edge.source_port}'. "
                        f"Available outputs: {[f'{o.value}_out' for o in source_outputs]}"
                    )

                if target_io_type not in target_inputs:
                    raise ValueError(
                        f"Node {target_node._id} does not have input port '{edge.target_port}'. "
                        f"Available inputs: {[f'{i.value}_in' for i in target_inputs]}"
                    )

            # Validate graph structure using the centralized validator
            validate_pipeline_graph(trigger_data, workers_data or [], edges_data)

        return pipeline
