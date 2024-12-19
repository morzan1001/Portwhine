#!/usr/bin/env python3
from fastapi import FastAPI, HTTPException, APIRouter
from typing import List, Union, Dict, Any

from api.models.trigger import IPAddressTrigger, CertstreamTrigger, WorkerConfig
from utils.logger import LoggingModule
from utils.elasticsearch import get_elasticsearch_connection
from docker.manager import DockerManager

router = APIRouter()
logger = LoggingModule()
docker_manager = DockerManager()

class TriggerHandler:
    def __init__(self):
        self.es_client = get_elasticsearch_connection()
        if not self.es_client:
            raise HTTPException(status_code=500, detail="Could not connect to Elasticsearch")
        self.index_name = "regex-store"

    async def lifespan(self, app: FastAPI):
        try:
            if not self.es_client.indices.exists(index=self.index_name):
                self.es_client.indices.create(index=self.index_name)
            yield
        except Exception as e:
            logger.error(f"Error during startup: {e}")
            raise HTTPException(status_code=500, detail=f"Error during startup: {str(e)}")

    def handle_ip_trigger(self, trigger: IPAddressTrigger):
        self.process_worker(trigger.worker)

    def handle_certstream_trigger(self, trigger: CertstreamTrigger):
        try:
            if not self.es_client.exists(index=self.index_name, id=trigger.regex):
                self.es_client.index(index=self.index_name, id=trigger.regex, body={"regex": trigger.regex})
            self.process_worker(trigger.worker)
        except Exception as e:
            logger.error(f"Error handling certstream trigger: {e}")
            raise HTTPException(status_code=500, detail=f"Error handling certstream trigger: {str(e)}")

    def process_worker(self, worker_config: WorkerConfig):
        logger.info(f"Processing worker: {worker_config.worker_name}")
        docker_manager.start_container(
            container_name=worker_config.worker_name,
            image_name="worker_image",  
            command="worker_command"
        )
        for child in worker_config.children:
            self.process_worker(child)

@router.post("/trigger")
async def trigger_endpoint(trigger: Union[IPAddressTrigger, CertstreamTrigger]):
    handler = TriggerHandler()
    if isinstance(trigger, IPAddressTrigger):
        handler.handle_ip_trigger(trigger)
    elif isinstance(trigger, CertstreamTrigger):
        handler.handle_certstream_trigger(trigger)
    return {"status": "success"}

@router.get("/triggers", response_model=List[str])
async def get_triggers():
    return ["IPAddressTrigger", "CertstreamTrigger"]

@router.get("/trigger/{name}", response_model=Dict[str, Any])
async def get_trigger_config(name: str):
    if name == "IPAddressTrigger":
        return {
            "description": "Trigger that accepts a list of IP addresses.",
            "example": IPAddressTrigger(
                ip_addresses=["192.168.1.1", "10.0.0.1"],
                worker=WorkerConfig(worker_name="IPWorker", children=[])
            ).dict()
        }
    elif name == "CertstreamTrigger":
        return {
            "description": "Trigger that accepts a regex pattern.",
            "example": CertstreamTrigger(
                regex=".*example.*",
                worker=WorkerConfig(worker_name="CertstreamWorker", children=[])
            ).dict()
        }
    else:
        raise HTTPException(status_code=404, detail="Trigger not found")