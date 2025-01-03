#!/usr/bin/env python3
import json
from elasticsearch import NotFoundError
from fastapi import HTTPException
from api.models.pipeline import Pipeline
from api.models.types import NodeStatus
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule
from utils.redis import get_redis_connection

class PipelineHandler:
    def __init__(self):
        self.logger = LoggingModule.get_logger()
        self.redis_client = get_redis_connection(db=1)
        self.es_client = get_elasticsearch_connection()

    def handle_pipeline_start(self, pipeline_id: str):
        try:
            # Retrieve the pipeline structure from Elasticsearch
            result = self.es_client.get(index="pipelines", id=pipeline_id)
            pipeline = Pipeline(**result["_source"])

            container_name = str(pipeline.trigger._id)
            task = {
                "pipeline_id": pipeline_id,
                "action": "start",
                "container_name": container_name,
                "image_name": pipeline.trigger.image_name,
                "environment": {"PIPELINE_ID": pipeline_id}
            }
            self.redis_client.rpush("container_queue", json.dumps(task))
            self.logger.info(f"Queued start for container {container_name}")

            # Create a new database index with the UUID of the pipeline
            if not self.es_client.indices.exists(index=pipeline_id):
                self.es_client.indices.create(index=pipeline_id)
            self.logger.info(f"Created new index {pipeline_id} for pipeline results.")

        except NotFoundError:
            raise HTTPException(status_code=404, detail="Pipeline not found")
        except Exception as e:
            self.logger.error(f"Error handling pipeline start: {e}")
            raise HTTPException(status_code=500, detail="Internal Server Error")

    def handle_pipeline_stop(self, pipeline_id: str):
        try:
            # Retrieve the pipeline structure from Elasticsearch
            result = self.es_client.get(index="pipelines", id=pipeline_id)
            pipeline = Pipeline(**result["_source"])

            # Stop the trigger container
            self.logger.debug(f"Queueing stop for trigger container with ID: {pipeline.trigger._id}")
            task = {
                "pipeline_id": pipeline_id,
                "action": "stop",
                "container_name": str(pipeline.trigger._id)
            }
            self.redis_client.rpush("container_queue", json.dumps(task))

            # Identify and stop the worker containers
            for worker in pipeline.worker:
                self.logger.debug(f"Queueing stop for worker container with ID: {worker._id}")
                task = {
                    "pipeline_id": pipeline_id,
                    "action": "stop",
                    "container_name": str(worker._id)
                }
                self.redis_client.rpush("container_queue", json.dumps(task))

            self.logger.info(f"Queued stop for all containers for pipeline {pipeline_id}.")

        except NotFoundError:
            raise HTTPException(status_code=404, detail="Pipeline not found")
        except Exception as e:
            self.logger.error(f"Error stopping pipeline: {e}")
            raise HTTPException(status_code=500, detail=f"Error stopping pipeline: {str(e)}")

    @staticmethod
    def update_status(pipeline_id: str, node_id: str, status: NodeStatus):
        try:
            es_client = get_elasticsearch_connection()
            logger = LoggingModule.get_logger()
            # Retrieve the pipeline structure from Elasticsearch
            result = es_client.get(index="pipelines", id=pipeline_id)
            pipeline = Pipeline(**result["_source"])

            logger.debug(f"Updating status for node {node_id} in pipeline {pipeline_id} to {status}")

            if node_id == str(pipeline.trigger._id):
                pipeline.trigger._status = status
            else:
                for worker in pipeline.worker:
                    if node_id == str(worker._id):
                        worker._status = status
                        break

            # Check if any trigger or worker is in error state
            is_any_error = (
                pipeline.trigger._status == NodeStatus.ERROR or
                any(worker._status == NodeStatus.ERROR for worker in pipeline.worker)
            )

            # Check if any trigger or worker is running
            is_any_running = (
                pipeline.trigger._status == NodeStatus.RUNNING or
                any(worker._status == NodeStatus.RUNNING for worker in pipeline.worker)
            )

            # Update pipeline status
            if is_any_error:
                pipeline_status = NodeStatus.ERROR
            else:
                pipeline_status = NodeStatus.RUNNING if is_any_running else NodeStatus.STOPPED
            pipeline._status = pipeline_status

            # Save updated pipeline status to Elasticsearch
            es_client.index(index="pipelines", id=pipeline_id, body=pipeline.ser_model())
            logger.info(f"Updated status for pipeline {pipeline_id} to {pipeline_status}")

        except NotFoundError:
            raise HTTPException(status_code=404, detail="Pipeline not found")
        except Exception as e:
            logger = LoggingModule.get_logger()
            logger.error(f"Error updating status: {e}")
            raise HTTPException(status_code=500, detail=f"Error updating status: {str(e)}")
        
    def cleanup_containers(self, node_id: str):
        try:
            self.logger.debug(f"Cleaning up all containers for worker {node_id}")
            task = {
                    "container_name": node_id,
                    "action": "cleanup",
                }
            self.redis_client.rpush("container_queue", json.dumps(task))
            self.logger.info(f"All containers for node {node_id} have been removed.")
        except Exception as e:
            self.logger.error(f"Error cleaning up containers for node {node_id}: {e}")
            raise