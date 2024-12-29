#!/usr/bin/env python3
import json
from fastapi import HTTPException
from api.models.pipeline import Pipeline
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule
from utils.redis import get_redis_connection

class PipelineHandler:
    def __init__(self):
        self.es_client = get_elasticsearch_connection()
        self.redis_client = get_redis_connection(db=1)
        self.logger = LoggingModule.get_logger()

    def handle_pipeline_start(self, pipeline_id: str):
        try:
            # Retrieve the pipeline structure from Elasticsearch
            result = self.es_client.get(index="pipelines", id=pipeline_id)
            pipeline = Pipeline(**result["_source"])

            container_name = str(pipeline.trigger.id)
            task = {
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

        except Exception as e:
            self.logger.error(f"Error handling pipeline start: {e}")
            raise HTTPException(status_code=500, detail="Internal Server Error")

    def handle_pipeline_stop(self, pipeline_id: str):
        try:
            # Retrieve the pipeline structure from Elasticsearch
            result = self.es_client.get(index="pipelines", id=pipeline_id)
            pipeline = Pipeline(**result["_source"])

            # Stop the trigger container
            self.logger.debug(f"Queueing stop for trigger container with ID: {pipeline.trigger.id}")
            task = {
                "action": "stop",
                "container_name": str(pipeline.trigger.id)
            }
            self.redis_client.rpush("container_queue", json.dumps(task))

            # Identify and stop the worker containers
            for worker in pipeline.worker:
                self.logger.debug(f"Queueing stop for worker container with ID: {worker.id}")
                task = {
                    "action": "stop",
                    "container_name": str(worker.id)
                }
                self.redis_client.rpush("container_queue", json.dumps(task))

            self.logger.info(f"Queued stop for all containers for pipeline {pipeline_id}.")

        except Exception as e:
            self.logger.error(f"Error stopping pipeline: {e}")
            raise HTTPException(status_code=500, detail=f"Error stopping pipeline: {str(e)}")