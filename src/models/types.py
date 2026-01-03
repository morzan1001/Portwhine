#!/usr/bin/env python3
from enum import Enum


class InputOutputType(str, Enum):
    """Data types that can flow between nodes."""

    HTTP = "http"
    IP = "ip"


class NodeStatus(str, Enum):
    """Runtime status of a node or pipeline."""

    PENDING = "Pending"
    STARTING = "Starting"
    RUNNING = "Running"
    PAUSED = "Paused"
    STOPPED = "Stopped"
    COMPLETED = "Completed"
    RESTARTING = "Restarting"
    OOM_KILLED = "OOMKilled"
    DEAD = "Dead"
    ERROR = "Error"
    UNKNOWN = "Unknown"

    @classmethod
    def is_active(cls, status: "NodeStatus") -> bool:
        """Check if the status indicates the node is actively processing."""
        return status in (cls.RUNNING, cls.STARTING, cls.RESTARTING)

    @classmethod
    def is_error(cls, status: "NodeStatus") -> bool:
        """Check if the status indicates an error condition."""
        return status in (cls.ERROR, cls.OOM_KILLED, cls.DEAD)

    @classmethod
    def is_stopped(cls, status: "NodeStatus") -> bool:
        """Check if the status indicates the node is stopped."""
        return status in (cls.STOPPED, cls.PAUSED, cls.COMPLETED)


class NodeType(str, Enum):
    """
    Fundamental type of node in a pipeline.

    - TRIGGER: Nodes that initiate pipeline execution (no inputs)
    - WORKER: Nodes that process data (have inputs and outputs)
    """

    TRIGGER = "trigger"
    WORKER = "worker"


class WorkerCategory(str, Enum):
    """
    Categories for organizing worker nodes in the UI.

    Only applies to workers - triggers don't have categories.
    """

    SCANNER = "scanner"
    ANALYZER = "analyzer"
    UTILITY = "utility"
    OUTPUT = "output"


class FieldType(str, Enum):
    """Supported field types for node configuration."""

    STRING = "string"
    INTEGER = "integer"
    FLOAT = "float"
    BOOLEAN = "boolean"
    SELECT = "select"
    MULTISELECT = "multiselect"
    IP_ADDRESS = "ip_address"
    IP_LIST = "ip_list"
    REGEX = "regex"
    PORT_RANGE = "port_range"
