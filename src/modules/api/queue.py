#!/usr/bin/env python3
import re
import threading
import time
import json
from api.models.types import NodeStatus
from utils.logger import LoggingModule
from utils.redis import get_redis_connection
from api.docker.manager import DockerManager

logger = LoggingModule.get_logger()
redis_client = get_redis_connection(db=1)
redis_container_health_client = get_redis_connection(db=2)
docker_manager = DockerManager()

def process_queue():
    while True:
        _, message = redis_client.blpop("container_queue")
        task = json.loads(message)
        action = task["action"]
        container_name = task.get("container_name")
        pipeline_id = task.get("pipeline_id")

        # Strip instance number to get clean node (trigger or worker) id
        node_id = re.sub(r'_instance_\d+$', '', container_name)

        if action == "start":
            image_name = task.get("image_name")
            command = task.get("command")
            environment = task.get("environment", {})
            logger.info(f"Starting container {container_name}")
            container = docker_manager.start_container(
                container_name=container_name,
                image_name=image_name,
                command=command,
                environment=environment
            )
            if container:
                redis_container_health_client.hset("active_containers", container_name, pipeline_id) 
        elif action == "stop":
            logger.info(f"Stopping container {container_name}")
            docker_manager.stop_container(container_name)
        elif action == "cleanup":
            logger.info(f"Cleaning up containers for container {container_name}")
            redis_container_health_client.hdel("active_containers", container_name)
            docker_manager.cleanup_containers(container_name)

            # Delete all Redis keys for the container
            keys_to_delete = redis_container_health_client.keys(f"*{container_name}*")
            for key in keys_to_delete:
                redis_container_health_client.delete(key)
                logger.info(f"Deleted Redis key: {key}")

        time.sleep(1)  # Optional: Sleep to prevent tight loop

def start_queue_thread():
    queue_thread = threading.Thread(target=process_queue)
    queue_thread.daemon = True
    queue_thread.start()
    return queue_thread