#!/usr/bin/env python3
"""
Centralized validation utilities for Pydantic models.

This module provides reusable validators and validation helpers
for ensuring data integrity across the Portwhine pipeline system.
"""
import re
import uuid
from typing import Any, Dict, List, Optional, Set

from models.types import InputOutputType


# =============================================================================
# REGEX PATTERNS
# =============================================================================

# Valid pipeline/node name pattern (alphanumeric, spaces, hyphens, underscores)
NAME_PATTERN = re.compile(r"^[\w\s\-]{1,100}$", re.UNICODE)

# Valid port range patterns for nmap
PORT_RANGE_PATTERN = re.compile(r"^(-p-|-p\d+(-\d+)?(,\d+(-\d+)?)*|--top-ports\s+\d+)$")

# Valid resolution pattern (e.g., "1920x1080")
RESOLUTION_PATTERN = re.compile(r"^\d{3,5}x\d{3,5}$")

# Valid Docker image name
DOCKER_IMAGE_PATTERN = re.compile(
    r"^[a-z0-9][a-z0-9._-]*(/[a-z0-9][a-z0-9._-]*)*(:[a-z0-9][a-z0-9._-]*)?$", re.IGNORECASE
)

# IPv4 pattern
IPV4_PATTERN = re.compile(
    r"^((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)(/\d{1,2})?$"
)

# IPv6 pattern (simplified)
IPV6_PATTERN = re.compile(r"^([0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}(/\d{1,3})?$")


# =============================================================================
# VALIDATION HELPERS
# =============================================================================


def validate_name(name: str, field_name: str = "name") -> str:
    """
    Validate a name field (pipeline name, node name, etc.).

    Args:
        name: The name to validate
        field_name: Name of the field for error messages

    Returns:
        The validated (stripped) name

    Raises:
        ValueError: If the name is invalid
    """
    if not name:
        raise ValueError(f"{field_name} cannot be empty")

    name = name.strip()

    if len(name) < 1:
        raise ValueError(f"{field_name} cannot be empty after trimming")

    if len(name) > 100:
        raise ValueError(f"{field_name} cannot exceed 100 characters")

    if not NAME_PATTERN.match(name):
        raise ValueError(f"{field_name} can only contain letters, numbers, spaces, " "hyphens, and underscores")

    return name


def validate_uuid(value: Any, field_name: str = "id") -> uuid.UUID:
    """
    Validate and convert a value to UUID.

    Args:
        value: The value to validate (string or UUID)
        field_name: Name of the field for error messages

    Returns:
        A valid UUID object

    Raises:
        ValueError: If the value is not a valid UUID
    """
    if isinstance(value, uuid.UUID):
        return value

    if isinstance(value, str):
        try:
            return uuid.UUID(value)
        except (ValueError, AttributeError) as exc:
            raise ValueError(f"{field_name} must be a valid UUID, got: {value}") from exc

    raise ValueError(f"{field_name} must be a UUID or UUID string, got: {type(value)}")


def validate_regex_pattern(pattern: str, field_name: str = "regex") -> str:
    """
    Validate that a string is a valid regex pattern.

    Args:
        pattern: The regex pattern to validate
        field_name: Name of the field for error messages

    Returns:
        The validated pattern

    Raises:
        ValueError: If the pattern is invalid
    """
    if not pattern:
        raise ValueError(f"{field_name} cannot be empty")

    try:
        re.compile(pattern)
    except re.error as e:
        raise ValueError(f"{field_name} is not a valid regex pattern: {e}") from e

    return pattern


def validate_port_range(ports: str, field_name: str = "ports") -> str:
    """
    Validate an nmap-style port range specification.

    Valid formats:
    - "-p-" (all ports)
    - "-p80" (single port)
    - "-p80,443" (multiple ports)
    - "-p1-1000" (port range)
    - "-p80,443,8000-9000" (mixed)
    - "--top-ports 100"

    Args:
        ports: The port specification to validate
        field_name: Name of the field for error messages

    Returns:
        The validated port specification

    Raises:
        ValueError: If the specification is invalid
    """
    if not ports:
        raise ValueError(f"{field_name} cannot be empty")

    ports = ports.strip()

    # Check for common valid patterns
    if ports == "-p-":
        return ports

    if ports.startswith("--top-ports"):
        parts = ports.split()
        if len(parts) == 2 and parts[1].isdigit():
            return ports
        raise ValueError(f"{field_name}: invalid --top-ports format")

    if ports.startswith("-p"):
        port_spec = ports[2:]
        # Validate each port/range
        for part in port_spec.split(","):
            if "-" in part:
                start, end = part.split("-", 1)
                if not (start.isdigit() and end.isdigit()):
                    raise ValueError(f"{field_name}: invalid port range '{part}'")
                if int(start) > int(end):
                    raise ValueError(f"{field_name}: invalid range {start}-{end}")
                if int(end) > 65535:
                    raise ValueError(f"{field_name}: port {end} exceeds 65535")
            else:
                if not part.isdigit():
                    raise ValueError(f"{field_name}: invalid port '{part}'")
                if int(part) > 65535:
                    raise ValueError(f"{field_name}: port {part} exceeds 65535")
        return ports

    raise ValueError(f"{field_name}: must start with '-p' or '--top-ports'")


def validate_resolution(resolution: str, field_name: str = "resolution") -> str:
    """
    Validate a screen resolution string.

    Args:
        resolution: The resolution to validate (e.g., "1920x1080")
        field_name: Name of the field for error messages

    Returns:
        The validated resolution

    Raises:
        ValueError: If the resolution is invalid
    """
    if not resolution:
        raise ValueError(f"{field_name} cannot be empty")

    resolution = resolution.strip().lower()

    if not RESOLUTION_PATTERN.match(resolution):
        raise ValueError(f"{field_name} must be in format WIDTHxHEIGHT (e.g., '1920x1080')")

    width, height = map(int, resolution.split("x"))

    if width < 100 or height < 100:
        raise ValueError(f"{field_name}: minimum resolution is 100x100")

    if width > 7680 or height > 4320:
        raise ValueError(f"{field_name}: maximum resolution is 7680x4320 (8K)")

    return resolution


def validate_wordlist_path(path: str, field_name: str = "wordlist") -> str:
    """
    Validate a wordlist file path.

    Args:
        path: The path to validate
        field_name: Name of the field for error messages

    Returns:
        The validated path

    Raises:
        ValueError: If the path is invalid
    """
    if not path:
        raise ValueError(f"{field_name} cannot be empty")

    path = path.strip()

    # Must be an absolute path
    if not path.startswith("/"):
        raise ValueError(f"{field_name} must be an absolute path starting with '/'")

    # Check for path traversal attempts
    if ".." in path:
        raise ValueError(f"{field_name} cannot contain '..' (path traversal)")

    # Check for potentially dangerous characters
    dangerous_chars = [";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">"]
    for char in dangerous_chars:
        if char in path:
            raise ValueError(f"{field_name} cannot contain '{char}'")

    return path


def validate_nmap_arguments(args: str, field_name: str = "arguments") -> str:
    """
    Validate nmap command-line arguments.

    Args:
        args: The nmap arguments to validate
        field_name: Name of the field for error messages

    Returns:
        The validated arguments

    Raises:
        ValueError: If the arguments are invalid or dangerous
    """
    if not args:
        return ""

    args = args.strip()

    # Blacklist dangerous options
    dangerous_options = [
        "--script-args-file",  # Could read arbitrary files
        "--datadir",  # Could access arbitrary directories
        "-iL",  # Input from file
        "-oN",
        "-oX",
        "-oG",
        "-oA",  # Output to files (we control this)
        "--resume",  # Could access arbitrary files
    ]

    args_lower = args.lower()
    for opt in dangerous_options:
        if opt.lower() in args_lower:
            raise ValueError(f"{field_name}: option '{opt}' is not allowed")

    # Check for shell injection attempts
    dangerous_chars = [";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">", "\n", "\r"]
    for char in dangerous_chars:
        if char in args:
            raise ValueError(f"{field_name}: character '{char}' is not allowed")

    return args


def validate_custom_nmap_command(command: Optional[str], field_name: str = "custom_command") -> Optional[str]:
    """
    Validate a custom nmap command.

    Args:
        command: The custom nmap command to validate
        field_name: Name of the field for error messages

    Returns:
        The validated command or None

    Raises:
        ValueError: If the command is invalid
    """
    if not command:
        return None

    command = command.strip()

    # Must start with 'nmap'
    if not command.startswith("nmap "):
        raise ValueError(f"{field_name} must start with 'nmap '")

    # Must contain the target placeholder
    if "{{target}}" not in command:
        raise ValueError(f"{field_name} must contain '{{{{target}}}}' placeholder")

    # Must output XML to stdout
    if "-oX -" not in command and "-oX-" not in command:
        raise ValueError(f"{field_name} must include '-oX -' for XML output to stdout")

    # Check for shell injection
    dangerous_chars = [";", "&", "|", "$", "`", "(", ")", "{", "}", "<", ">", "\n", "\r"]
    # Allow {{ and }} for placeholder
    temp_command = command.replace("{{target}}", "")
    for char in dangerous_chars:
        if char in temp_command:
            raise ValueError(f"{field_name}: character '{char}' is not allowed")

    return command


def validate_file_extensions(extensions: Optional[str], field_name: str = "extensions") -> Optional[str]:
    """
    Validate a comma-separated list of file extensions.

    Args:
        extensions: The extensions string to validate
        field_name: Name of the field for error messages

    Returns:
        The validated extensions or None

    Raises:
        ValueError: If the extensions are invalid
    """
    if not extensions:
        return None

    extensions = extensions.strip()

    # Split and validate each extension
    ext_list = [e.strip().lower() for e in extensions.split(",")]

    for ext in ext_list:
        if not ext:
            continue
        # Extension should be alphanumeric only
        if not ext.isalnum():
            raise ValueError(f"{field_name}: extension '{ext}' contains invalid characters")
        if len(ext) > 10:
            raise ValueError(f"{field_name}: extension '{ext}' is too long (max 10 chars)")

    return ",".join(ext_list)


# =============================================================================
# EDGE VALIDATION
# =============================================================================


def validate_edge_connectivity(
    source_outputs: List[InputOutputType],
    target_inputs: List[InputOutputType],
    source_port: Optional[str] = None,
    target_port: Optional[str] = None,
) -> bool:
    """
    Validate that an edge can connect two nodes based on their port types.

    Args:
        source_outputs: Output types of the source node
        target_inputs: Input types of the target node
        source_port: Optional specific source port
        target_port: Optional specific target port

    Returns:
        True if the connection is valid

    Raises:
        ValueError: If the connection is invalid
    """
    source_set = set(source_outputs)
    target_set = set(target_inputs)

    # Check if there's any compatible type
    compatible_types = source_set & target_set

    if not compatible_types:
        raise ValueError(
            f"Incompatible connection: source outputs {[t.value for t in source_outputs]} "
            f"cannot connect to target inputs {[t.value for t in target_inputs]}"
        )

    # If specific ports are specified, validate them
    if source_port:
        # Port ID format: {type}_out (e.g., "http_out", "ip_out")
        expected_suffix = "_out"
        if not source_port.endswith(expected_suffix):
            raise ValueError(f"Invalid source port format: {source_port}")

        port_type = source_port[: -len(expected_suffix)]
        try:
            source_type = InputOutputType(port_type)
            if source_type not in source_outputs:
                raise ValueError(f"Source port '{source_port}' not available on source node")
        except ValueError as exc:
            raise ValueError(f"Unknown data type in source port: {port_type}") from exc

    if target_port:
        expected_suffix = "_in"
        if not target_port.endswith(expected_suffix):
            raise ValueError(f"Invalid target port format: {target_port}")

        port_type = target_port[: -len(expected_suffix)]
        try:
            target_type = InputOutputType(port_type)
            if target_type not in target_inputs:
                raise ValueError(f"Target port '{target_port}' not available on target node")
        except ValueError as exc:
            raise ValueError(f"Unknown data type in target port: {port_type}") from exc

    # If both ports specified, ensure they match
    if source_port and target_port:
        source_type = source_port.replace("_out", "")
        target_type = target_port.replace("_in", "")
        if source_type != target_type:
            raise ValueError(f"Port type mismatch: cannot connect {source_port} to {target_port}")

    return True


# =============================================================================
# PIPELINE GRAPH VALIDATION
# =============================================================================


def validate_pipeline_graph(
    trigger: Optional[Any],
    workers: List[Any],
    edges: List[Any],
) -> None:
    """
    Validate the pipeline graph structure.

    Checks:
    - All edge sources and targets exist
    - No self-loops
    - No duplicate edges
    - Graph is connected (all workers reachable from trigger)
    - No cycles

    Args:
        trigger: Trigger node object (with _id attribute) or None
        workers: List of worker node objects (with _id attribute)
        edges: List of Edge objects (with source and target attributes)

    Raises:
        ValueError: If the graph structure is invalid
    """
    trigger_id = trigger._id if trigger else None
    worker_ids = {w._id for w in workers} if workers else set()

    if not trigger_id and not worker_ids:
        return  # Empty pipeline is valid

    if worker_ids and not trigger_id:
        raise ValueError("Pipeline has workers but no trigger")

    all_nodes = worker_ids.copy()
    if trigger_id:
        all_nodes.add(trigger_id)

    # Validate edges
    edge_set = set()
    adjacency: Dict[uuid.UUID, Set[uuid.UUID]] = {node: set() for node in all_nodes}

    for edge in edges:
        source = edge.source if hasattr(edge, "source") else edge.get("source")
        target = edge.target if hasattr(edge, "target") else edge.get("target")

        if isinstance(source, str):
            source = uuid.UUID(source)
        if isinstance(target, str):
            target = uuid.UUID(target)

        # Check nodes exist
        if source not in all_nodes:
            raise ValueError(f"Edge source {source} not found in pipeline")
        if target not in all_nodes:
            raise ValueError(f"Edge target {target} not found in pipeline")

        # Check self-loop
        if source == target:
            raise ValueError(f"Self-loop detected on node {source}")

        # Check duplicate
        edge_key = (source, target)
        if edge_key in edge_set:
            raise ValueError(f"Duplicate edge from {source} to {target}")
        edge_set.add(edge_key)

        adjacency[source].add(target)

    # Check connectivity (all workers reachable from trigger)
    if trigger_id and worker_ids:
        reachable = set()
        queue = [trigger_id]

        while queue:
            node = queue.pop(0)
            if node in reachable:
                continue
            reachable.add(node)
            queue.extend(adjacency[node])

        unreachable = worker_ids - reachable
        if unreachable:
            raise ValueError(f"Workers not reachable from trigger: {[str(w) for w in unreachable]}")

    # Check for cycles using DFS
    def has_cycle(start: uuid.UUID, visited: Set[uuid.UUID], rec_stack: Set[uuid.UUID]) -> bool:
        visited.add(start)
        rec_stack.add(start)

        for neighbor in adjacency[start]:
            if neighbor not in visited:
                if has_cycle(neighbor, visited, rec_stack):
                    return True
            elif neighbor in rec_stack:
                return True

        rec_stack.remove(start)
        return False

    visited: Set[uuid.UUID] = set()
    for node in all_nodes:
        if node not in visited:
            if has_cycle(node, visited, set()):
                raise ValueError("Pipeline contains a cycle")
