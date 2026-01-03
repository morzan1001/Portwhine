#!/usr/bin/env python3
"""
WebSocket router for real-time pipeline status updates.
"""
import json
import asyncio
from typing import Dict, Set
from fastapi import APIRouter, WebSocket, WebSocketDisconnect

from utils.logger import LoggingModule
from utils.elasticsearch import get_elasticsearch_connection
from models.pipeline import Pipeline
from models.types import NodeStatus

router = APIRouter()
logger = LoggingModule.get_logger()


class ConnectionManager:
    """Manages WebSocket connections for pipeline status updates."""
    
    def __init__(self):
        # Map of pipeline_id -> set of connected WebSockets
        self.active_connections: Dict[str, Set[WebSocket]] = {}
        # Map of WebSocket -> set of subscribed pipeline_ids
        self.subscriptions: Dict[WebSocket, Set[str]] = {}
    
    async def connect(self, websocket: WebSocket):
        """Accept a new WebSocket connection."""
        await websocket.accept()
        self.subscriptions[websocket] = set()
        logger.debug(f"WebSocket connected: {websocket.client}")
    
    def disconnect(self, websocket: WebSocket):
        """Handle WebSocket disconnection."""
        # Remove from all pipeline subscriptions
        if websocket in self.subscriptions:
            for pipeline_id in self.subscriptions[websocket]:
                if pipeline_id in self.active_connections:
                    self.active_connections[pipeline_id].discard(websocket)
                    if not self.active_connections[pipeline_id]:
                        del self.active_connections[pipeline_id]
            del self.subscriptions[websocket]
        logger.debug(f"WebSocket disconnected: {websocket.client}")
    
    def subscribe(self, websocket: WebSocket, pipeline_id: str):
        """Subscribe a WebSocket to a pipeline's updates."""
        if pipeline_id not in self.active_connections:
            self.active_connections[pipeline_id] = set()
        self.active_connections[pipeline_id].add(websocket)
        self.subscriptions[websocket].add(pipeline_id)
        logger.debug(f"WebSocket subscribed to pipeline: {pipeline_id}")
    
    def unsubscribe(self, websocket: WebSocket, pipeline_id: str):
        """Unsubscribe a WebSocket from a pipeline's updates."""
        if pipeline_id in self.active_connections:
            self.active_connections[pipeline_id].discard(websocket)
        if websocket in self.subscriptions:
            self.subscriptions[websocket].discard(pipeline_id)
        logger.debug(f"WebSocket unsubscribed from pipeline: {pipeline_id}")
    
    async def broadcast_to_pipeline(self, pipeline_id: str, message: dict):
        """Send a message to all WebSockets subscribed to a pipeline."""
        if pipeline_id not in self.active_connections:
            return
        
        disconnected = set()
        for websocket in self.active_connections[pipeline_id]:
            try:
                await websocket.send_json(message)
            except Exception as e:
                logger.warning(f"Failed to send to WebSocket: {e}")
                disconnected.add(websocket)
        
        # Clean up disconnected sockets
        for websocket in disconnected:
            self.disconnect(websocket)
    
    async def send_personal(self, websocket: WebSocket, message: dict):
        """Send a message to a specific WebSocket."""
        try:
            await websocket.send_json(message)
        except Exception as e:
            logger.warning(f"Failed to send personal message: {e}")


manager = ConnectionManager()


async def get_pipeline_status(pipeline_id: str) -> dict:
    """Fetch current pipeline status from Elasticsearch."""
    es_client = get_elasticsearch_connection()
    if not es_client:
        return {"error": "Database connection failed"}
    
    try:
        result = es_client.get(index="pipelines", id=pipeline_id)
        pipeline_data = result["_source"]
        
        # Extract status information for the pipeline and all nodes
        status_info = {
            "pipeline_id": pipeline_id,
            "pipeline_status": pipeline_data.get("status", NodeStatus.STOPPED),
            "nodes": []
        }
        
        # Add trigger status
        if pipeline_data.get("trigger"):
            trigger_data = pipeline_data["trigger"]
            trigger_name = list(trigger_data.keys())[0] if trigger_data else None
            if trigger_name:
                node_data = trigger_data[trigger_name]
                status_info["nodes"].append({
                    "id": node_data.get("id"),
                    "type": trigger_name,
                    "status": node_data.get("status", NodeStatus.STOPPED),
                    "is_trigger": True,
                    "instance_health": node_data.get("instanceHealth")
                })
        
        # Add worker statuses
        for worker in pipeline_data.get("worker", []):
            worker_name = list(worker.keys())[0] if worker else None
            if worker_name:
                node_data = worker[worker_name]
                status_info["nodes"].append({
                    "id": node_data.get("id"),
                    "type": worker_name,
                    "status": node_data.get("status", NodeStatus.STOPPED),
                    "is_trigger": False,
                    "instance_health": node_data.get("instanceHealth"),
                    "number_of_instances": node_data.get("numberOfInstances", 0)
                })
        
        return status_info
    
    except Exception as e:
        logger.error(f"Error fetching pipeline status: {e}")
        return {"error": str(e)}


async def poll_pipeline_status(websocket: WebSocket, pipeline_id: str):
    """Poll pipeline status and send updates.
    
    Uses adaptive polling intervals:
    - 5 seconds when pipeline is stopped
    - 2 seconds when pipeline is running
    """
    last_status = None
    
    while True:
        try:
            current_status = await get_pipeline_status(pipeline_id)
            
            # Only send if status changed
            if current_status != last_status:
                await manager.send_personal(websocket, {
                    "type": "status_update",
                    "data": current_status
                })
                last_status = current_status
            
            # Adaptive polling interval based on pipeline status
            pipeline_status = current_status.get("pipeline_status", "Stopped")
            if pipeline_status in ["Running", "Starting"]:
                await asyncio.sleep(2)  # Poll more frequently when running
            else:
                await asyncio.sleep(5)  # Poll less frequently when stopped
        except asyncio.CancelledError:
            break
        except Exception as e:
            logger.error(f"Error in status polling: {e}")
            await asyncio.sleep(interval)


@router.websocket("/ws/pipeline/{pipeline_id}")
async def websocket_pipeline_status(websocket: WebSocket, pipeline_id: str):
    """
    WebSocket endpoint for real-time pipeline status updates.
    
    Clients connect to this endpoint to receive live updates about:
    - Pipeline status changes
    - Node status changes (running, error, completed, etc.)
    - Instance health updates
    
    Message format (sent to client):
    {
        "type": "status_update",
        "data": {
            "pipeline_id": "...",
            "pipeline_status": "Running",
            "nodes": [
                {
                    "id": "node-uuid",
                    "type": "NmapWorker",
                    "status": "Running",
                    "is_trigger": false,
                    "instance_health": [...],
                    "number_of_instances": 3
                }
            ]
        }
    }
    """
    await manager.connect(websocket)
    manager.subscribe(websocket, pipeline_id)
    
    # Start polling task
    poll_task = asyncio.create_task(poll_pipeline_status(websocket, pipeline_id))
    
    try:
        # Send initial status
        initial_status = await get_pipeline_status(pipeline_id)
        await manager.send_personal(websocket, {
            "type": "status_update",
            "data": initial_status
        })
        
        # Keep connection alive and handle incoming messages
        while True:
            try:
                data = await websocket.receive_json()
                
                # Handle client commands
                if data.get("type") == "ping":
                    await manager.send_personal(websocket, {"type": "pong"})
                
                elif data.get("type") == "subscribe":
                    new_pipeline_id = data.get("pipeline_id")
                    if new_pipeline_id:
                        manager.subscribe(websocket, new_pipeline_id)
                        status = await get_pipeline_status(new_pipeline_id)
                        await manager.send_personal(websocket, {
                            "type": "status_update",
                            "data": status
                        })
                
                elif data.get("type") == "unsubscribe":
                    old_pipeline_id = data.get("pipeline_id")
                    if old_pipeline_id:
                        manager.unsubscribe(websocket, old_pipeline_id)
                
            except WebSocketDisconnect:
                break
            except Exception as e:
                logger.warning(f"Error processing WebSocket message: {e}")
    
    finally:
        poll_task.cancel()
        manager.disconnect(websocket)


@router.websocket("/ws/pipelines")
async def websocket_all_pipelines(websocket: WebSocket):
    """
    WebSocket endpoint for monitoring all pipelines.
    Useful for the pipeline list view to show live status indicators.
    """
    await manager.connect(websocket)
    
    try:
        while True:
            try:
                data = await websocket.receive_json()
                
                if data.get("type") == "subscribe":
                    pipeline_id = data.get("pipeline_id")
                    if pipeline_id:
                        manager.subscribe(websocket, pipeline_id)
                        status = await get_pipeline_status(pipeline_id)
                        await manager.send_personal(websocket, {
                            "type": "status_update",
                            "data": status
                        })
                
                elif data.get("type") == "unsubscribe":
                    pipeline_id = data.get("pipeline_id")
                    if pipeline_id:
                        manager.unsubscribe(websocket, pipeline_id)
                
                elif data.get("type") == "ping":
                    await manager.send_personal(websocket, {"type": "pong"})
            
            except WebSocketDisconnect:
                break
            except Exception as e:
                logger.warning(f"Error in pipelines WebSocket: {e}")
    
    finally:
        manager.disconnect(websocket)
