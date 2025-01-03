from typing import List, Dict, Any
from fastapi import APIRouter
from api.models.trigger import IPAddressTrigger, CertstreamTrigger, TriggerConfig
from utils.logger import LoggingModule
from api.docs.trigger_docs import trigger_summaries, trigger_descriptions

router = APIRouter()
logger = LoggingModule.get_logger()

trigger_classes = [IPAddressTrigger, CertstreamTrigger]

@router.get("/trigger",
            response_model=List[str],
            summary=trigger_summaries["get_triggers"],
            description=trigger_descriptions["get_triggers"])
async def get_triggers():
    return [cls.__name__ for cls in trigger_classes]

@router.get("/trigger/{name}",
    response_model=Dict[str, Any],
    summary=trigger_summaries["get_trigger_config"],
    description=trigger_descriptions["get_trigger_config"],
)
async def get_trigger_config(name: str):
    for cls in trigger_classes:
        if cls.__name__ == name:
            # Create example instance with predefined example values
            if cls == CertstreamTrigger:
                example_instance = cls(regex="^example\\.com$")
            elif cls == IPAddressTrigger:
                example_instance = cls(ip_addresses=["192.168.0.1"])
            else:
                example_instance = cls()
            # Clean up the docstring to remove leading/trailing whitespace and newlines
            description = cls.__doc__.strip() if cls.__doc__ else "No description available"
            return {
                "description": description,
                "example": example_instance.model_dump()
            }
    return {"error": "Trigger not found"}