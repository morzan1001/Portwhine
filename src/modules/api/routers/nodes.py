#!/usr/bin/env python3
"""
Router for node definitions - provides metadata about available nodes.
"""
from typing import List, Optional
from fastapi import APIRouter, HTTPException

from models.node_definition import NodeDefinition
from models.node_registry import NodeRegistry

router = APIRouter()


@router.get(
    "/nodes",
    response_model=List[NodeDefinition],
    summary="Get all node definitions",
    description="Returns metadata for all available nodes (triggers and workers). "
                "Includes port definitions, configuration fields, and UI properties."
)
async def get_all_nodes():
    """Get all available node definitions."""
    return NodeRegistry.get_all()


# Note: Specific routes must be defined BEFORE the generic {node_id} route
# to prevent FastAPI from matching "triggers" or "workers" as a node_id

@router.get(
    "/nodes/triggers",
    response_model=List[NodeDefinition],
    summary="Get all trigger nodes",
    description="Returns all available trigger node definitions."
)
async def get_trigger_nodes():
    """Get all trigger node definitions."""
    return NodeRegistry.get_triggers()


@router.get(
    "/nodes/workers",
    response_model=List[NodeDefinition],
    summary="Get all worker nodes",
    description="Returns all available worker node definitions."
)
async def get_worker_nodes():
    """Get all worker node definitions."""
    return NodeRegistry.get_workers()


@router.get(
    "/nodes/category/{category}",
    response_model=List[NodeDefinition],
    summary="Get nodes by category",
    description="Returns all nodes in a specific category (trigger, scanner, analyzer, utility, output)."
)
async def get_nodes_by_category(category: str):
    """Get all nodes in a category."""
    nodes = NodeRegistry.get_by_category(category)
    if not nodes:
        raise HTTPException(status_code=404, detail=f"No nodes found in category '{category}'")
    return nodes


@router.get(
    "/nodes/{node_id}",
    response_model=NodeDefinition,
    summary="Get node definition by ID",
    description="Returns metadata for a specific node type."
)
async def get_node(node_id: str):
    """Get a specific node definition."""
    node = NodeRegistry.get_by_id(node_id)
    if not node:
        raise HTTPException(status_code=404, detail=f"Node '{node_id}' not found")
    return node
