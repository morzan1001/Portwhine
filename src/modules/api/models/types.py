#!/usr/bin/env python3
from enum import Enum

class InputOutputType(str, Enum):
    HTTP = "http"
    IP = "ip"

class NodeStatus(str, Enum):
    RUNNING = "Running"
    STOPPED = "Stopped"
    ERROR = "Error"

