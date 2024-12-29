worker_responses = {
    200: {
        "description": "Successful response with worker configuration",
        "content": {
            "application/json": {
                "example": {
                    "description": "Example configuration of a worker.",
                    "example": {
                        "NmapWorker": {
                            "id": "6950db58-a608-4cc8-969a-2058f4a30367",
                            "input": ["ip"],
                            "output": ["ip", "http"],
                            "children": []
                        }
                    }
                }
            }
        }
    },
    404: {
        "description": "Worker not found",
        "content": {
            "application/json": {
                "example": {"error": "Worker not found"}
            }
        }
    }
}

worker_summaries = {
    "get_workers": "Get all worker names",
    "get_worker_config": "Get the configuration of a specific worker"
}

worker_descriptions = {
    "get_workers": "Retrieve a list of all available worker names.",
    "get_worker_config": "Retrieve the configuration of a specific worker by name."
}