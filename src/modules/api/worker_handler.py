#!/usr/bin/env python3
import json
from api.models.worker import WorkerConfig
from api.models.job_payload import JobPayload
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

    def start_worker(self, pipeline_id: str, worker: WorkerConfig, job_payload: JobPayload):
        input_fields = worker.input

        try:
            payload_json = job_payload.model_dump_json()
            worker_id = str(worker._id)
            payload_key = f"worker:{worker_id}:payloads"


            # Check if the payload already exists in the Redis set
            if self.redis_block_client.sismember(payload_key, payload_json):
                self.logger.info(f"Payload for worker {worker_id} already exists in Redis.")
                return

            # Save Worker ID and JOB_Payload to Redis set with a TTL of 24 hours
            self.redis_block_client.sadd(payload_key, payload_json)
            self.redis_block_client.expire(payload_key, 86400)
            self.logger.info(f"Saved JOB_Payload for worker {worker_id} to Redis set with a TTL of 24 hours")
            # Count the number of already started instances
            instance_key = f"worker:{worker_id}:instances"
            existing_instances = int(self.redis_queue_client.get(instance_key) or 0)
            instance_number = existing_instances + 1

            # Update the numberOfInstances for the worker
            self.es_client.update(index="pipeline_index", id=pipeline_id, body={
                "script": {
                    "source": "for (int i = 0; i < ctx._source.worker.size(); i++) { if (ctx._source.worker[i].id == params.worker_id) { ctx._source.worker[i].numberOfInstances = params.instance_number; break; } }",
                    "lang": "painless",
                    "params": {
                        "worker_id": worker_id,
                        "instance_number": instance_number
                    }
                }
            })
            
            container_name = f"{worker_id}_instance_{instance_number}"

            self.redis_queue_client.set(instance_key, instance_number)

            # Start the worker container
            task = {
                "pipeline_id": pipeline_id,
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

    def stop_worker(self, pipeline_id: str, worker_id: str):
        try:
            # Stop the worker container
            container_name = f"{worker_id}_instance"
            task = {
                "pipeline_id": pipeline_id,
                "action": "stop",
                "container_name": container_name
            }
            self.redis_queue_client.rpush("container_queue", json.dumps(task))
            self.logger.info(f"Task for stopping container {container_name} added to queue")

        except Exception as e:
            self.logger.error(f"Error stopping worker {worker_id}: {e}")
            raise