#!/usr/bin/env python3
import json
from models.worker import WorkerConfig
from models.job_payload import JobPayload
from utils.redis import get_redis_connection
from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule
from utils.helper import json_serial

class WorkerHandler:
    def __init__(self):
        self.redis_block_client = get_redis_connection(db=0)
        self.redis_queue_client = get_redis_connection(db=1)
        self.es_client = get_elasticsearch_connection()
        self.logger = LoggingModule.get_logger()

    def start_worker(self, pipeline_id: str, worker: WorkerConfig, job_payload: JobPayload, run_id: str):
        input_fields = worker.input

        try:
            payload_json = job_payload.model_dump_json()
            worker_id = str(worker._id)
            
            # We don't need to check for existing payload in Redis for now, 
            # as we are moving to a run-based approach.
            # But we can keep it if we want to avoid duplicate work within a run?
            # For now, let's assume Orchestrator handles logic.

            # Count the number of already started instances
            instance_key = f"worker:{worker_id}:instances"
            existing_instances = int(self.redis_queue_client.get(instance_key) or 0)
            instance_number = existing_instances + 1
            
            container_name = f"{worker_id}_instance_{instance_number}"

            self.redis_queue_client.set(instance_key, instance_number)

            # Serialize worker config
            worker_config = worker.model_dump_json()

            # Start the worker container
            task = {
                "pipeline_id": pipeline_id,
                "action": "start",
                "container_name": container_name,
                "image_name": worker.image_name,
                "environment": {
                    "JOB_PAYLOAD": payload_json,
                    "PIPELINE_ID": pipeline_id,
                    "WORKER_ID": worker_id,
                    "INSTANCE_NAME": container_name,
                    "RUN_ID": run_id,
                    "WORKER_CONFIG": worker_config
                }
            }
            self.redis_queue_client.rpush("container_queue", json.dumps(task, default=json_serial))
            self.logger.info(f"Task for starting container {container_name} added to queue for run {run_id}")

        except Exception as e:
            self.logger.error(f"Error starting worker {worker_id}: {e}")
            raise

    def start_trigger(self, pipeline_id: str, trigger, run_id: str):
        try:
            trigger_id = str(trigger._id)
            container_name = f"{trigger_id}_instance"
            
            # Serialize trigger config for the container
            trigger_config = trigger.model_dump_json()

            task = {
                "pipeline_id": pipeline_id,
                "action": "start",
                "container_name": container_name,
                "image_name": trigger.image_name,
                "environment": {
                    "PIPELINE_ID": pipeline_id,
                    "WORKER_ID": trigger_id, # Triggers act as workers too
                    "INSTANCE_NAME": container_name,
                    "RUN_ID": run_id,
                    "TRIGGER_CONFIG": trigger_config
                }
            }
            self.redis_queue_client.rpush("container_queue", json.dumps(task, default=json_serial))
            self.logger.info(f"Task for starting trigger {container_name} added to queue for run {run_id}")
        except Exception as e:
            self.logger.error(f"Error starting trigger {trigger._id}: {e}")
            raise

    def cleanup_worker(self, pipeline_id: str, container_name: str):
        try:
            # Cleanup the worker container
            task = {
                "pipeline_id": pipeline_id,
                "action": "cleanup",
                "container_name": container_name
            }
            self.redis_queue_client.rpush("container_queue", json.dumps(task))
            self.logger.info(f"Task for cleaning up container {container_name} added to queue")

        except Exception as e:
            self.logger.error(f"Error cleaning up worker {container_name}: {e}")
            raise