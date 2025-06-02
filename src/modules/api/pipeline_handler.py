#!/usr/bin/env python3
import json
import os
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

            if not pipeline.trigger:
                raise HTTPException(status_code=400, detail="Pipeline cannot be started without a trigger")

            if pipeline._status != NodeStatus.STOPPED:
                raise HTTPException(status_code=400, detail="Pipeline is already running")

            container_name = str(pipeline.trigger._id)
            task = {
                "pipeline_id": pipeline_id,
                "action": "start",
                "container_name": container_name,
                "image_name": pipeline.trigger.image_name,
                "environment": {"PIPELINE_ID": pipeline_id, 
                                "LOG_LEVEL": os.getenv("LOG_LEVEL", "INFO")},
            }
            self.redis_client.rpush("container_queue", json.dumps(task))
            self.logger.info(f"Queued start for container {container_name}")

            # Create a new database index with the UUID of the pipeline
            if not self.es_client.indices.exists(index=pipeline_id):
                self.es_client.indices.create(index=pipeline_id)
            self.logger.info(f"Created new index {pipeline_id} for pipeline results.")

        except HTTPException as http_exc:
            raise http_exc
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

            if pipeline._status == NodeStatus.STOPPED:
                raise HTTPException(status_code=400, detail="Pipeline is already stopped")

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

        except HTTPException as http_exc:
            raise http_exc
        except NotFoundError:
            raise HTTPException(status_code=404, detail="Pipeline not found")
        except Exception as e:
            self.logger.error(f"Error stopping pipeline: {e}")
            raise HTTPException(status_code=500, detail=f"Error stopping pipeline: {str(e)}")
        
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
    
    def cleanup_instance_count(self, worker_id: str, pipeline_id: str):
        # Cleanup worker fields in Elasticsearch
        script = {
            "source": """
                for (int i = 0; i < ctx._source.worker.size(); i++) {
                    if (ctx._source.worker[i].id == params.worker_id) {
                        ctx._source.worker[i].numberOfInstances = 0;
                        ctx._source.worker[i].instanceHealth = None;
                        break;
                    }
                }
            """,
            "lang": "painless",
            "params": {
                "worker_id": worker_id
            }
        }
        self.es_client.update(index="pipeline_index", id=pipeline_id, body={"script": script})
        self.logger.info(f"Cleaned up worker {worker_id} in pipeline {pipeline_id}")