#!/usr/bin/env python3
from enum import Enum

class InputOutputType(str, Enum):
    HTTP = "http"
    IP = "ip"

class NodeStatus(str, Enum):
    STATUS = "Status"
    PENDING = "Pending"
    STARTING = "Starting"
    RUNNING = "Running"
    PAUSED = "Paused"
    STOPPED = "Stopped"
    COMPLETED = "Completed"
    RESTARTING = "Restarting"
    OOMKILLED = "OOMKilled"
    DEAD = "Dead"
    PID = "Pid"
    EXITCODE = "ExitCode"
    ERROR = "Error"
    STARTEDAT = "StartedAt"
    FINISHEDAT = "FinishedAt"