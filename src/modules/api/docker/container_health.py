#!/usr/bin/env python3
import threading
import time
import re
import requests
import docker
from models.types import NodeStatus
from utils.logger import LoggingModule
from utils.redis import get_redis_connection
from utils.elasticsearch import get_elasticsearch_connection

logger = LoggingModule.get_logger()
redis_client = get_redis_connection(db=2)
es_client = get_elasticsearch_connection()

def get_container_health(container_name) -> NodeStatus:
    client = docker.client.from_env()
    try:
        container = client.containers.get(container_name)
        docker_status = container.attrs.get("State", {}).get("Status")
        
        if docker_status != "running":
            return NodeStatus.STOPPED if docker_status == "exited" else NodeStatus.ERROR

        # If running, check internal health endpoint
        try:
            networks = container.attrs.get("NetworkSettings", {}).get("Networks", {})
            # Assume first network or look for specific one
            ip_address = None
            for net_name, net_info in networks.items():
                if net_info.get("IPAddress"):
                    ip_address = net_info.get("IPAddress")
                    break
            
            if not ip_address:
                logger.warning(f"Container {container_name} running but no IP found")
                return NodeStatus.RUNNING # Fallback to Docker status

            response = requests.get(f"http://{ip_address}:8000/health", timeout=2)
            if response.status_code == 200:
                data = response.json()
                return data.get("status", NodeStatus.RUNNING)
            else:
                logger.warning(f"Container {container_name} health check failed: {response.status_code}")
                return NodeStatus.ERROR
        except Exception as e:
            logger.warning(f"Container {container_name} health check exception: {e}")
            # If we can't reach the health endpoint but docker says running, 
            # it might be starting up or really stuck.
            # Let's return RUNNING for now to avoid false positives during startup,
            # or maybe STARTING if we had that state.
            return NodeStatus.RUNNING

    except docker.client.errors.NotFound:
        return "container not found"
    
def update_pipeline_status(pipeline_id):
    try:
        pipeline = es_client.get(index="pipeline_index", id=pipeline_id)["_source"]
        
        # Collecting all statuses of triggers and workers
        trigger_statuses = [trigger["status"] for trigger in pipeline.get("trigger", [])]
        worker_statuses = [worker["status"] for worker in pipeline.get("worker", [])]
        all_statuses = trigger_statuses + worker_statuses

        # Determine the global status of the pipeline based on the status of the triggers and workers
        if all(status == NodeStatus.PAUSED for status in all_statuses):
            global_status = NodeStatus.PAUSED
        elif any(status == NodeStatus.ERROR or status == NodeStatus.OOMKILLED for status in all_statuses):
            global_status = NodeStatus.ERROR
        elif any(status == NodeStatus.RUNNING for status in all_statuses):
            global_status = NodeStatus.RUNNING
        else:
            global_status = NodeStatus.STOPPED

        # Updating the global status of the pipeline in Elasticsearch
        script = {
            "source": "ctx._source.status = params.global_status",
            "lang": "painless",
            "params": {
                "global_status": global_status
            }
        }
        es_client.update(index="pipeline_index", id=pipeline_id, body={"script": script})
        logger.info(f"Updated global status of pipeline {pipeline_id} to {global_status}")
    except Exception as e:
        logger.error(f"Error updating global status of pipeline {pipeline_id}: {e}")

def update_worker_status(pipeline_id, node_id):
    try:
        workers = es_client.get(index="pipeline_index", id=pipeline_id)["_source"]
        worker = next((w for w in workers["worker"] if w["id"] == node_id), None)
        
        if not worker:
            logger.error(f"Worker {node_id} not found in pipeline {pipeline_id}")
            return

        # Determine the global status of the worker based on the instance status
        instance_statuses = [ih["health"] for ih in worker.get("instanceHealth", [])]
        if all(status == NodeStatus.PAUSED for status in instance_statuses):
            global_status = NodeStatus.PAUSED
        elif any(status == NodeStatus.ERROR or status == NodeStatus.OOMKILLED for status in instance_statuses):
            global_status = NodeStatus.ERROR
        elif any(status == NodeStatus.RUNNING for status in instance_statuses):
            global_status = NodeStatus.RUNNING
        else:
            global_status = NodeStatus.STOPPED

        # Updating the global status of the worker in Elasticsearch
        script = {
            "source": """
                for (int i = 0; i < ctx._source.worker.size(); i++) {
                    if (ctx._source.worker[i].id == params.node_id) {
                        ctx._source.worker[i].status = params.global_status;
                        break;
                    }
                }
            """,
            "lang": "painless",
            "params": {
                "node_id": node_id,
                "global_status": global_status
            }
        }
        es_client.update(index="pipeline_index", id=pipeline_id, body={"script": script})
        logger.info(f"Updated global status of worker {node_id} in pipeline {pipeline_id} to {global_status}")

        # Update the global status of the pipeline
        update_pipeline_status(pipeline_id)
    except Exception as e:
        logger.error(f"Error updating global status of worker {node_id} in pipeline {pipeline_id}: {e}")


def update_instance_health(pipeline_id, node_id, instance_id, health_status):
    try:
        # Update the instance health status directly in Elasticsearch
        script = {
            "source": """
                for (int i = 0; i < ctx._source.worker.size(); i++) {
                    if (ctx._source.worker[i].id == params.node_id) {
                        boolean instanceFound = false;
                        for (int j = 0; j < ctx._source.worker[i].instanceHealth.size(); j++) {
                            if (ctx._source.worker[i].instanceHealth[j].number == params.instance_id) {
                                ctx._source.worker[i].instanceHealth[j].health = params.health_status;
                                instanceFound = true;
                                break;
                            }
                        }
                        if (!instanceFound) {
                            ctx._source.worker[i].instanceHealth.add(params.instance_health);
                        }
                        break;
                    }
                }
            """,
            "lang": "painless",
            "params": {
                "node_id": node_id,
                "instance_id": instance_id,
                "health_status": health_status,
                "instance_health": {
                    "number": instance_id,
                    "health": health_status
                }
            }
        }
        es_client.update(index="pipeline_index", id=pipeline_id, body={"script": script})
        logger.info(f"Updated health status of instance {instance_id} for node {node_id} in pipeline {pipeline_id} to {health_status}")

        # Update the global status of the worker based on the instance health status
        update_worker_status(pipeline_id, node_id)
    except Exception as e:
        logger.error(f"Error updating instance health status in Elasticsearch: {e}")

def update_trigger_health(pipeline_id, node_id, health_status):
    try:
        # Update the trigger health status directly in Elasticsearch
        script = {
            "source": """
                for (int i = 0; i < ctx._source.trigger.size(); i++) {
                    if (ctx._source.trigger[i].id == params.node_id) {
                        ctx._source.trigger[i].status = params.health_status;
                        break;
                    }
                }
            """,
            "lang": "painless",
            "params": {
                "node_id": node_id,
                "health_status": health_status
            }
        }
        es_client.update(index="pipeline_index", id=pipeline_id, body={"script": script})
        logger.info(f"Updated health status of trigger {node_id} in pipeline {pipeline_id} to {health_status}")

        # Update the global status of the pipeline
        update_pipeline_status(pipeline_id)
    except Exception as e:
        logger.error(f"Error updating trigger health status in Elasticsearch: {e}")

def monitor_health():
    logger.info("Health monitoring thread started.")
    while True:
        container_names = redis_client.hkeys("active_containers")
        for container_name in container_names:
            container_name = container_name.decode("utf-8")
            pipeline_id = redis_client.hget("active_containers", container_name).decode("utf-8")
            health_status = get_container_health(container_name)
            logger.info(f"Container {container_name} Status: {health_status}")

            if "instance" in container_name:
                # Strip instance number to get clean node (trigger or worker) id
                node_id = re.sub(r'_instance_\d+$', '', container_name)
                # Extract instance number from container name
                instance_id_match = re.search(r'_instance_(\d+)$', container_name)
                instance_id = int(instance_id_match.group(1)) if instance_id_match else 0

                # Update Elasticsearch with the instance health status
                if es_client:
                    update_instance_health(pipeline_id, node_id, instance_id, health_status)
                else:
                    logger.error("Elasticsearch client is not available")
            else:
                # Update Elasticsearch with the trigger health status
                if es_client:
                    update_trigger_health(pipeline_id, container_name, health_status)
                else:
                    logger.error("Elasticsearch client is not available")

        time.sleep(2)  # Optional: Sleep to prevent tight loop

def start_container_health_thread():
    queue_thread = threading.Thread(target=monitor_health)
    queue_thread.daemon = True
    queue_thread.start()
    return queue_thread