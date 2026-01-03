#!/usr/bin/env python3
import ipaddress
import uuid
from typing import Any, Dict, Set

def json_serial(obj):
    """JSON serializer for objects not serializable by default json code"""
    if isinstance(obj, uuid.UUID):
        return str(obj)
    raise TypeError(f"Type {type(obj)} not serializable")

def is_private_ip(ip):
    try:
        ip_obj = ipaddress.ip_address(ip)
        return ip_obj.is_private
    except ValueError:
        return False


# Runtime-only fields that should not be stored in the database
# These are computed at runtime from ClassVars, not persisted configuration
RUNTIME_ONLY_FIELDS: Set[str] = {"status", "inputs", "outputs", "input", "output", "instanceHealth"}


def strip_runtime_fields(data: Dict[str, Any]) -> Dict[str, Any]:
    """
    Remove runtime-only fields from pipeline data before loading into Pydantic models.
    
    This is needed because:
    1. ser_model() includes runtime fields (status, inputs, outputs) for API responses
    2. Legacy data in Elasticsearch may contain these fields
    3. Our models use extra="forbid" for strict validation
    
    Args:
        data: Raw pipeline data from Elasticsearch or API request
        
    Returns:
        Cleaned data with runtime fields removed from trigger and worker configs
    """
    if not isinstance(data, dict):
        return data
    
    result = data.copy()
    
    # Clean trigger config
    trigger = result.get("trigger")
    if isinstance(trigger, dict):
        cleaned_trigger = {}
        for trigger_type, trigger_config in trigger.items():
            if isinstance(trigger_config, dict):
                cleaned_trigger[trigger_type] = {
                    k: v for k, v in trigger_config.items() 
                    if k not in RUNTIME_ONLY_FIELDS
                }
            else:
                cleaned_trigger[trigger_type] = trigger_config
        result["trigger"] = cleaned_trigger
    
    # Clean worker configs
    workers = result.get("worker")
    if isinstance(workers, list):
        cleaned_workers = []
        for worker_entry in workers:
            if isinstance(worker_entry, dict):
                cleaned_entry = {}
                for worker_type, worker_config in worker_entry.items():
                    if isinstance(worker_config, dict):
                        cleaned_entry[worker_type] = {
                            k: v for k, v in worker_config.items()
                            if k not in RUNTIME_ONLY_FIELDS
                        }
                    else:
                        cleaned_entry[worker_type] = worker_config
                cleaned_workers.append(cleaned_entry)
            else:
                cleaned_workers.append(worker_entry)
        result["worker"] = cleaned_workers
    
    return result