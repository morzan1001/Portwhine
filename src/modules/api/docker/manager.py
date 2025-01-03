#!/usr/bin/env python3
import os
import docker
from utils.logger import LoggingModule

class DockerManager:
    def __init__(self):
        self.client = docker.from_env()
        self.logger = LoggingModule.get_logger()

    def start_container(self, container_name, image_name, command=None, network_name="portwhine", **kwargs):
        try:
            self.logger.debug(f"Attempting to start container: {container_name}")
            # Set environment variables for database connection
            environment = kwargs.get('environment', {})
            environment.update({
                "DATABASE_HOST": os.getenv("DATABASE_HOST", "http://elasticsearch:9200"),
                "DATABASE_USER": os.getenv("DATABASE_USER", "elastic"),
                "DATABASE_PASSWORD": os.getenv("DATABASE_PASSWORD", "changeme")
            })
            kwargs['environment'] = environment

            # Set network
            kwargs['network'] = network_name

            container = self.client.containers.run(
                image_name,
                name=container_name,
                command=command,
                detach=True,
                **kwargs
            )
            self.logger.info(f"Started container {container_name} using image {image_name} with command '{command}' in network '{network_name}'.")
            return container
        except docker.errors.APIError as e:
            self.logger.error(f"Error starting container {container_name}: {e}")
            return None

    def stop_container(self, container_name):
        try:
            self.logger.debug(f"Attempting to stop container: {container_name}")
            container = self.client.containers.get(container_name)
            if container.status == "running":
                container.stop()
                container.remove()
                self.logger.info(f"Stopped and removed {container_name} container.")
            else:
                self.logger.warning(f"{container_name} container is not running.")
        except docker.errors.NotFound:
            self.logger.warning(f"{container_name} container not found.")
        except docker.errors.APIError as e:
            self.logger.error(f"Error stopping container {container_name}: {e}")

    def remove_container(self, container_name):
        try:
            self.logger.debug("Attempting to remove container: %s", container_name)
            container = self.client.containers.get(container_name)
            container.remove(force=True)
            self.logger.info("Removed container %s.", container_name)
        except docker.errors.NotFound:
            self.logger.warning("Container %s not found.", container_name)
        except docker.errors.APIError as e:
            self.logger.error("Error removing container %s: %s", container_name, e)

    def cleanup_containers(self, node_id: str):
        try:
            self.logger.debug("Attempting to clean up containers for node: %s", node_id)
            containers = self.client.containers.list(all=True, filters={"name": node_id})
            for container in containers:
                container.remove(force=True)
                self.logger.info("Removed container %s for worker %s.", container.name, node_id)
        except docker.errors.APIError as e:
            self.logger.error("Error cleaning up containers for worker %s: %s", node_id, e)