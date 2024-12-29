#!/usr/bin/env python3
from enum import Enum

class InputOutputType(str, Enum):
    HTTP = "http"
    IP = "ip"