#!/usr/bin/env python3
import ipaddress
import uuid

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