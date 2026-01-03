#!/usr/bin/env python3
"""
Node Registry - Central registry for all available node definitions.

This module auto-generates NodeDefinition objects from Worker and Trigger classes,
avoiding duplicate data definitions. It extracts metadata from:
- Class docstrings for descriptions
- Class variables (inputs, outputs, image_name, category, display_name)
- Pydantic Field definitions for configuration fields
"""
from typing import Dict, List, Type, Any, get_origin, get_args, Union
from pydantic import BaseModel

from models.node_definition import NodeDefinition, PortDefinition, FieldDefinition
from models.types import InputOutputType, WorkerCategory, FieldType, NodeType
from models.worker import (
    WorkerConfig,
    FFUFWorker,
    HumbleWorker,
    ScreenshotWorker,
    TestSSLWorker,
    WebAppAnalyzerWorker,
    NmapWorker,
    ResolverWorker,
)
from models.trigger import TriggerConfig, IPAddressTrigger, CertstreamTrigger


# Node appearance configuration (colors and icons)
# These are UI-specific and don't belong in the model classes
NODE_APPEARANCE: Dict[str, Dict[str, str]] = {
    # Triggers
    "IPAddressTrigger": {"color": "#10B981", "icon": "network"},
    "CertstreamTrigger": {"color": "#F59E0B", "icon": "certificate"},
    # Scanners
    "NmapWorker": {"color": "#6366F1", "icon": "radar"},
    "FFUFWorker": {"color": "#EC4899", "icon": "search"},
    # Analyzers
    "HumbleWorker": {"color": "#14B8A6", "icon": "shield"},
    "TestSSLWorker": {"color": "#F97316", "icon": "lock"},
    "WebAppAnalyzerWorker": {"color": "#06B6D4", "icon": "code"},
    # Utilities
    "ScreenshotWorker": {"color": "#EF4444", "icon": "camera"},
    "ResolverWorker": {"color": "#8B5CF6", "icon": "dns"},
}


def _unwrap_optional(annotation: Any) -> Any:
    """Unwrap Optional types to get the inner type."""
    origin = get_origin(annotation)
    if origin is Union:
        args = get_args(annotation)
        non_none_args = [a for a in args if a is not type(None)]
        if non_none_args:
            return non_none_args[0]
    return annotation


def _is_ip_type(annotation: Any) -> bool:
    """Check if the annotation is an IP-related type."""
    annotation_str = str(annotation)
    return "IPvAny" in annotation_str or "Network" in annotation_str


def _python_type_to_field_type(annotation: Any) -> FieldType:
    """Convert Python type annotation to FieldType enum."""
    # Unwrap Optional types
    annotation = _unwrap_optional(annotation)
    origin = get_origin(annotation)

    # Handle List types
    if origin is list:
        args = get_args(annotation)
        if args and _is_ip_type(args[0]):
            return FieldType.IP_LIST
        return FieldType.STRING

    # Check for IP types
    if _is_ip_type(annotation):
        return FieldType.IP_ADDRESS

    # Map basic Python types to FieldType
    type_mapping = {
        str: FieldType.STRING,
        int: FieldType.INTEGER,
        float: FieldType.FLOAT,
        bool: FieldType.BOOLEAN,
    }

    return type_mapping.get(annotation, FieldType.STRING)


def _extract_config_fields(cls: Type[BaseModel], base_fields: set) -> List[FieldDefinition]:
    """Extract FieldDefinition objects from a Pydantic model's fields."""
    fields = []

    for name, field_info in cls.model_fields.items():
        if name in base_fields:
            continue

        # Get the annotation
        annotation = cls.__annotations__.get(name, str)

        # Determine if required
        is_required = field_info.is_required()

        # Get default value
        default_value = None
        if field_info.default is not None and not isinstance(field_info.default, type):
            default_value = field_info.default

        # Convert type
        field_type = _python_type_to_field_type(annotation)

        # Special case: check if field name contains 'regex'
        if "regex" in name.lower():
            field_type = FieldType.REGEX

        # Create label from field name
        label = name.replace("_", " ").title()

        fields.append(
            FieldDefinition(
                name=name,
                label=label,
                type=field_type,
                description=field_info.description or "",
                required=is_required,
                default=default_value,
                placeholder=str(default_value) if default_value else None,
            )
        )

    return fields


def _create_port_definitions(data_types: List[InputOutputType], is_input: bool) -> List[PortDefinition]:
    """Create PortDefinition objects from a list of data types."""
    ports = []
    suffix = "in" if is_input else "out"

    for data_type in data_types:
        port_id = f"{data_type.value}_{suffix}"
        label = f"{data_type.value.upper()} {'Input' if is_input else 'Output'}"

        ports.append(
            PortDefinition(
                id=port_id,
                label=label,
                data_type=data_type,
                description=f"{'Accepts' if is_input else 'Produces'} {data_type.value} data",
            )
        )

    return ports


def _clean_docstring(docstring: str | None) -> str:
    """Clean and format a docstring for use as description."""
    if not docstring:
        return ""
    # Remove leading/trailing whitespace and normalize
    lines = docstring.strip().split("\n")
    # Join and clean up extra whitespace
    return " ".join(line.strip() for line in lines if line.strip())


def _create_definition_from_worker(cls: Type[WorkerConfig]) -> NodeDefinition:
    """Create a NodeDefinition from a Worker class."""
    class_name = cls.__name__
    appearance = NODE_APPEARANCE.get(class_name, {"color": "#6366F1", "icon": "default"})

    # Base fields to exclude from config
    base_fields = {"gridPosition", "numberOfInstances", "instanceHealth"}

    return NodeDefinition(
        id=class_name,
        name=getattr(cls, "display_name", class_name.replace("Worker", "")),
        description=_clean_docstring(cls.__doc__) or f"Worker: {class_name}",
        node_type=NodeType.WORKER,
        category=getattr(cls, "category", WorkerCategory.UTILITY),
        icon=appearance["icon"],
        color=appearance["color"],
        inputs=_create_port_definitions(cls.inputs, is_input=True),
        outputs=_create_port_definitions(cls.outputs, is_input=False),
        config_fields=_extract_config_fields(cls, base_fields),
        image_name=cls.image_name,
        supports_multiple_instances=True,
    )


def _create_definition_from_trigger(cls: Type[TriggerConfig]) -> NodeDefinition:
    """Create a NodeDefinition from a Trigger class."""
    class_name = cls.__name__
    appearance = NODE_APPEARANCE.get(class_name, {"color": "#10B981", "icon": "trigger"})

    # Base fields to exclude from config
    base_fields = {"gridPosition"}

    return NodeDefinition(
        id=class_name,
        name=getattr(cls, "display_name", class_name.replace("Trigger", "")),
        description=_clean_docstring(cls.__doc__) or f"Trigger: {class_name}",
        node_type=NodeType.TRIGGER,
        category=None,  # Triggers don't have categories
        icon=appearance["icon"],
        color=appearance["color"],
        inputs=[],  # Triggers have no inputs
        outputs=_create_port_definitions(cls.outputs, is_input=False),
        config_fields=_extract_config_fields(cls, base_fields),
        image_name=cls.image_name,
        supports_multiple_instances=False,
    )


class NodeRegistry:
    """
    Central registry for all node definitions.

    Auto-generates NodeDefinition objects from Worker and Trigger classes
    on first access. This avoids duplicate data definitions and ensures
    consistency between the models and their UI representations.
    """

    _definitions: Dict[str, NodeDefinition] = {}
    _initialized: bool = False

    # All registered worker classes
    WORKER_CLASSES: List[Type[WorkerConfig]] = [
        NmapWorker,
        ResolverWorker,
        FFUFWorker,
        HumbleWorker,
        ScreenshotWorker,
        TestSSLWorker,
        WebAppAnalyzerWorker,
    ]

    # All registered trigger classes
    TRIGGER_CLASSES: List[Type[TriggerConfig]] = [
        IPAddressTrigger,
        CertstreamTrigger,
    ]

    @classmethod
    def _init_definitions(cls):
        """Initialize all node definitions from registered classes."""
        if cls._initialized:
            return

        # Generate definitions from trigger classes
        for trigger_cls in cls.TRIGGER_CLASSES:
            definition = _create_definition_from_trigger(trigger_cls)
            cls._definitions[definition.id] = definition

        # Generate definitions from worker classes
        for worker_cls in cls.WORKER_CLASSES:
            definition = _create_definition_from_worker(worker_cls)
            cls._definitions[definition.id] = definition

        cls._initialized = True

    @classmethod
    def get_all(cls) -> List[NodeDefinition]:
        """Get all registered node definitions."""
        cls._init_definitions()
        return list(cls._definitions.values())

    @classmethod
    def get(cls, node_id: str) -> NodeDefinition:
        """
        Get a specific node definition by ID.

        Args:
            node_id: The node type identifier (class name)

        Returns:
            The NodeDefinition for the requested node type

        Raises:
            KeyError: If the node type is not registered
        """
        cls._init_definitions()
        if node_id not in cls._definitions:
            raise KeyError(f"Unknown node type: {node_id}")
        return cls._definitions[node_id]

    @classmethod
    def get_by_id(cls, node_id: str) -> NodeDefinition | None:
        """Get a specific node definition by ID, returns None if not found."""
        cls._init_definitions()
        return cls._definitions.get(node_id)

    @classmethod
    def get_by_category(cls, category: WorkerCategory) -> List[NodeDefinition]:
        """Get all worker node definitions in a specific category."""
        cls._init_definitions()
        return [d for d in cls._definitions.values() if d.category == category]

    @classmethod
    def get_triggers(cls) -> List[NodeDefinition]:
        """Get all trigger node definitions."""
        cls._init_definitions()
        return [d for d in cls._definitions.values() if d.node_type == NodeType.TRIGGER]

    @classmethod
    def get_workers(cls) -> List[NodeDefinition]:
        """Get all worker node definitions."""
        cls._init_definitions()
        return [d for d in cls._definitions.values() if d.node_type == NodeType.WORKER]

    @classmethod
    def exists(cls, node_id: str) -> bool:
        """Check if a node type is registered."""
        cls._init_definitions()
        return node_id in cls._definitions

    @classmethod
    def register_worker(cls, worker_cls: Type[WorkerConfig]) -> None:
        """
        Register a new worker class at runtime.

        Args:
            worker_cls: The worker class to register
        """
        cls._init_definitions()
        definition = _create_definition_from_worker(worker_cls)
        cls._definitions[definition.id] = definition
        cls.WORKER_CLASSES.append(worker_cls)

    @classmethod
    def register_trigger(cls, trigger_cls: Type[TriggerConfig]) -> None:
        """
        Register a new trigger class at runtime.

        Args:
            trigger_cls: The trigger class to register
        """
        cls._init_definitions()
        definition = _create_definition_from_trigger(trigger_cls)
        cls._definitions[definition.id] = definition
        cls.TRIGGER_CLASSES.append(trigger_cls)
