#!/usr/bin/env python3
import uuid
from typing import Any, Dict, List, Optional, Set
from pydantic import BaseModel, PrivateAttr, SerializeAsAny, model_serializer, model_validator
from api.models.trigger import TriggerConfig, CertstreamTrigger, IPAddressTrigger
from api.models.worker import ResolverWorker, WorkerConfig, FFUFWorker, HumbleWorker, ScreenshotWorker, TestSSLWorker, WebAppAnalyzerWorker, NmapWorker
from api.models.types import InputOutputType, NodeStatus
from utils.logger import LoggingModule

logger = LoggingModule.get_logger()

from api.models.edge import Edge

class Pipeline(BaseModel):
    _id: uuid.UUID = PrivateAttr(default_factory=uuid.uuid4)
    _status: str = PrivateAttr(default=NodeStatus.STOPPED)
    name: str
    trigger: Optional[SerializeAsAny[TriggerConfig]] = None
    worker: Optional[List[SerializeAsAny[WorkerConfig]]] = None
    edges: List[Edge] = []

    def __init__(self, **data: Any):
        super().__init__(**data)
        if 'id' in data:
            self._id = uuid.UUID(data['id'])
        if 'status' in data:
            self._status = NodeStatus(data['status'])

    @model_serializer
    def ser_model(self) -> dict[str, Any]:
        data = {
            "id": str(self._id),
            "status": self._status,
            "name": self.name,
            "trigger": self.trigger.ser_model() if self.trigger else None,
            "worker": [workers_data.ser_model() for workers_data in self.worker] if self.worker else [],
            "edges": [edge.model_dump() for edge in self.edges]
        }
        return data

    @model_validator(mode="before")
    def validate_trigger(cls, values: Dict[str, Any]) -> Dict[str, Any]:
        trigger_data = values.get("trigger")
        logger.debug(f"Validating trigger: {trigger_data}")
        if trigger_data == {}:
            values["trigger"] = None
        elif trigger_data:
            trigger_classes = [CertstreamTrigger, IPAddressTrigger]
            for trigger_cls in trigger_classes:
                key = trigger_cls.__name__
                if key in trigger_data:
                    values["trigger"] = trigger_cls(**trigger_data[key])
                    break
            else:
                raise ValueError("Invalid trigger data")
        return values

    @model_validator(mode="before")
    def validate_workers(cls, values: Dict[str, Any]) -> Dict[str, Any]:
        workers_data = values.get("worker")
        logger.debug(f"Validating workers: {workers_data}")
        if workers_data:
            worker_classes = [
                FFUFWorker, HumbleWorker, ScreenshotWorker, TestSSLWorker,
                WebAppAnalyzerWorker, NmapWorker, ResolverWorker
            ]
            def validate_worker(worker_data: Dict[str, Any]) -> WorkerConfig:
                for wcls in worker_classes:
                    key = wcls.__name__
                    if key in worker_data:
                        worker_dict = worker_data[key].copy()
                        worker_instance = wcls(**worker_dict)
                        return worker_instance
                raise ValueError("Invalid worker type")

            validated_workers = [validate_worker(worker_data) for worker_data in workers_data]
            values["worker"] = validated_workers
        return values

    @model_validator(mode="after")
    def validate_pipeline(cls, pipeline: 'Pipeline'):
        trigger_data = pipeline.trigger
        workers_data = pipeline.worker
        edges_data = pipeline.edges

        # Check whether both trigger and worker do not exist
        if not trigger_data and not workers_data:
            return pipeline

        # Check whether workers exist, but no trigger is present
        if workers_data and not trigger_data:
            raise ValueError("Workers cannot exist without a trigger")

        # Validate edges
        if edges_data:
            nodes = {worker._id: worker for worker in workers_data} if workers_data else {}
            if trigger_data:
                nodes[trigger_data._id] = trigger_data

            for edge in edges_data:
                source_node = nodes.get(edge.source)
                target_node = nodes.get(edge.target)

                if not source_node:
                    raise ValueError(f"Source node {edge.source} not found in pipeline")
                if not target_node:
                    raise ValueError(f"Target node {edge.target} not found in pipeline")

                source_output_set: Set[InputOutputType] = set(source_node.output)
                target_input_set: Set[InputOutputType] = set(target_node.input)

                if not source_output_set & target_input_set:
                    raise ValueError(f"Node {source_node._id} output does not match node {target_node._id} input")
        
        return pipeline

        return pipeline