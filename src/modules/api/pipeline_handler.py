#!/usr/bin/env python3
import json
from elasticsearch import NotFoundError
from fastapi import HTTPException
from api.orchestrator import Orchestrator
from models.pipeline import Pipeline
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule
from utils.redis import get_redis_connection
from utils.helper import strip_runtime_fields

class PipelineHandler:
    def __init__(self):
        self.logger = LoggingModule.get_logger()
        self.redis_client = get_redis_connection(db=1)
        self.es_client = get_elasticsearch_connection()
        self.orchestrator = Orchestrator()

    def handle_pipeline_start(self, pipeline_id: str):
        try:
            # Use Orchestrator to start the pipeline run
            self.orchestrator.start_pipeline(pipeline_id)
            self.logger.info("Pipeline %s started via Orchestrator.", pipeline_id)

        except HTTPException as http_exc:
            raise http_exc
        except NotFoundError as exc:
            raise HTTPException(status_code=404, detail="Pipeline not found") from exc
        except Exception as e:
            self.logger.error("Error handling pipeline start: %s", e)
            raise HTTPException(status_code=500, detail="Internal Server Error") from e

    def handle_pipeline_stop(self, pipeline_id: str):
        # Stopping logic might need to be updated to stop specific runs or all runs.
        # For now, we keep the logic to stop all containers related to the pipeline ID,
        # which is a bit broad but safe.
        try:
            # Retrieve the pipeline structure from Elasticsearch
            result = self.es_client.get(index="pipelines", id=pipeline_id)
            clean_data = strip_runtime_fields(result["_source"])
            pipeline = Pipeline(**clean_data)

            # Stop the trigger container
            if pipeline.trigger:
                self.logger.debug("Queueing cleanup for trigger container with ID: %s", pipeline.trigger.id)
                task = {
                    "pipeline_id": pipeline_id,
                    "action": "cleanup",
                    "container_name": str(pipeline.trigger.id)
                }
                self.redis_client.rpush("container_queue", json.dumps(task))

            # Identify and stop the worker containers
            if pipeline.worker:
                for worker in pipeline.worker:
                    self.logger.debug("Queueing cleanup for worker container with ID: %s", worker.id)
                    task = {
                        "pipeline_id": pipeline_id,
                        "action": "cleanup",
                        "container_name": str(worker.id)
                    }
                    self.redis_client.rpush("container_queue", json.dumps(task))

            self.logger.info("Queued cleanup for all containers for pipeline %s.", pipeline_id)

        except HTTPException as http_exc:
            raise http_exc
        except NotFoundError as exc:
            raise HTTPException(status_code=404, detail="Pipeline not found") from exc
        except Exception as e:
            self.logger.error("Error stopping pipeline: %s", e)
            raise HTTPException(status_code=500, detail=f"Error stopping pipeline: {str(e)}") from e
        
    def cleanup_containers(self, node_id: str):
        try:
            self.logger.debug("Cleaning up all containers for worker %s", node_id)
            task = {
                    "container_name": node_id,
                    "action": "cleanup",
                }
            self.redis_client.rpush("container_queue", json.dumps(task))

            self.logger.info("All containers for node %s have been removed.", node_id)
        except Exception as e:
            self.logger.error("Error cleaning up containers for node %s: %s", node_id, e)
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