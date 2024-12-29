import threading
import time
import json
from utils.logger import LoggingModule
from utils.redis import get_redis_connection
from api.docker.manager import DockerManager

logger = LoggingModule.get_logger()
redis_client = get_redis_connection(db=1)
docker_manager = DockerManager()

def process_queue():
    while True:
        _, message = redis_client.blpop("container_queue")
        task = json.loads(message)
        action = task["action"]
        container_name = task["container_name"]

        if action == "start":
            image_name = task["image_name"]
            command = task.get("command")
            environment = task.get("environment", {})
            logger.info(f"Starting container {container_name}")
            docker_manager.start_container(
                container_name=container_name,
                image_name=image_name,
                command=command,
                environment=environment
            )
        elif action == "stop":
            logger.info(f"Stopping container {container_name}")
            docker_manager.stop_container(container_name)

        time.sleep(1)  # Optional: Sleep to prevent tight loop

def start_queue_thread():
    queue_thread = threading.Thread(target=process_queue)
    queue_thread.daemon = True
    queue_thread.start()
    return queue_thread