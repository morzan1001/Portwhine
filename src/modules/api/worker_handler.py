#!/usr/bin/env python3
import json
from api.models.worker import WorkerConfig
from api.models.job_payload import JobPayload
from utils.redis import get_redis_connection
from utils.logger import LoggingModule
from utils.helper import json_serial

class WorkerHandler:
    def __init__(self):
        self.redis_block_client = get_redis_connection(db=0)
        self.redis_queue_client = get_redis_connection(db=1)
        self.logger = LoggingModule.get_logger()

    def start_worker(self, pipeline_id: str, worker: WorkerConfig, job_payload: JobPayload):
        worker_id = str(worker.id)
        input_fields = worker.input

        try:
            # Convert JobPayload to dict
            payload_json = job_payload.model_dump_json()
            payload_dict = json.loads(payload_json)

            # Check if a worker with the same ID and payload already exists
            existing_payloads = self.redis_block_client.lrange(f"worker:{worker_id}:payloads", 0, -1)
            for existing_payload_json in existing_payloads:
                existing_payload = json.loads(existing_payload_json)
                self.logger.debug(f"Payload JSON: {payload_json}")
                self.logger.debug(f"Existing payload: {existing_payload}")
                if all(field is not None and payload_dict.get(field) == existing_payload.get(field) for field in input_fields):
                    self.logger.info(f"Worker {worker_id} with the same payload already running. Skipping start.")
                    return

            # Save Worker ID and JOB_Payload to Redis with a TTL of 24 hours
            self.redis_block_client.rpush(f"worker:{worker_id}:payloads", payload_json)
            self.redis_block_client.expire(f"worker:{worker_id}:payloads", 86400)
            self.logger.info(f"Saved JOB_Payload for worker {worker_id} to Redis with a TTL of 24 hours")

            # Count the number of already started instances
            instance_key = f"worker:{worker_id}:instances"
            existing_instances = int(self.redis_queue_client.get(instance_key) or 0)
            instance_number = existing_instances + 1
            container_name = f"{worker_id}_instance_{instance_number}"

            self.redis_queue_client.set(instance_key, instance_number)

            # Start the worker container
            task = {
                "action": "start",
                "container_name": container_name,
                "image_name": worker.image_name,
                "environment": {
                    "JOB_PAYLOAD": payload_json,
                    "PIPELINE_ID": pipeline_id,
                    "WORKER_ID": worker_id
                }
            }
            self.redis_queue_client.rpush("container_queue", json.dumps(task, default=json_serial))
            self.logger.info(f"Task for starting container {container_name} added to queue")

        except Exception as e:
            self.logger.error(f"Error starting worker {worker_id}: {e}")
            raise

    def stop_worker(self, worker_id: str):
        try:
            # Stop the worker container
            container_name = f"{worker_id}_instance"
            task = {
                "action": "stop",
                "container_name": container_name
            }
            self.redis_queue_client.rpush("container_queue", json.dumps(task))
            self.logger.info(f"Task for stopping container {container_name} added to queue")

        except Exception as e:
            self.logger.error(f"Error stopping worker {worker_id}: {e}")
            raise