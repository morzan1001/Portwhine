trigger_responses = {
    200: {
        "description": "Successful response with trigger configuration",
        "content": {
            "application/json": {
                "example": {
                    "description": "Trigger that accepts a regex pattern.",
                    "example": {
                        "CertstreamTrigger": {
                            "id": "6950db58-a608-4cc8-969a-2058f4a30367",
                            "regex": "^example\\.com$",
                            "output": ["http"]
                        }
                    }
                }
            }
        }
    },
    404: {
        "description": "Trigger not found",
        "content": {
            "application/json": {
                "example": {"error": "Trigger not found"}
            }
        }
    }
}

trigger_summaries = {
    "get_triggers": "Get all trigger names",
    "get_trigger_config": "Get the configuration of a specific trigger"
}

trigger_descriptions = {
    "get_triggers": "Retrieve a list of all available trigger names.",
    "get_trigger_config": "Retrieve the configuration of a specific trigger by name."
}