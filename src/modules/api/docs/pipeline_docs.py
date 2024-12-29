pipeline_responses = {
    200: {
        "description": "Successful response with pipeline configuration",
        "content": {
            "application/json": {
                "example": {
                    "id": "a21b0a59-4181-4e94-97d5-615ddf09dfab",
                    "name": "Test Pipeline",
                    "trigger": {
                        "CertstreamTrigger": {
                            "id": "6950db58-a608-4cc8-969a-2058f4a30367",
                            "regex": "^example\\.com$",
                            "output": ["http"]
                        }
                    },
                    "worker": [
                        {
                            "ResolverWorker": {
                                "id": "0c975de2-f0fd-42b5-8ef5-985661eebae3",
                                "input": ["http"],
                                "output": ["ip"],
                                "children": [
                                    {
                                        "NmapWorker": {
                                            "id": "1c975de2-f0fd-42b5-8ef5-985661eebae3",
                                            "input": ["ip"],
                                            "output": ["ip", "http"],
                                            "children": [
                                                {
                                                    "FFUFWorker": {
                                                        "id": "2c975de2-f0fd-42b5-8ef5-985661eebae3",
                                                        "input": ["http"],
                                                        "output": ["http"]
                                                    }
                                                }
                                            ]
                                        }
                                    }
                                ]
                            }
                        }
                    ]
                }
            }
        }
    },
    404: {
        "description": "Pipeline not found",
        "content": {
            "application/json": {
                "example": {"error": "Pipeline not found"}
            }
        }
    }
}

pipeline_summaries = {
    "create_pipeline": "Create a new pipeline",
    "get_pipeline": "Get a pipeline configuration",
    "get_all_pipelines": "Get all pipeline configurations",
    "update_pipeline": "Update a pipeline configuration",
    "delete_pipeline": "Delete a pipeline configuration",
    "start_pipeline": "Start a pipeline",
    "stop_pipeline": "Stop a pipeline"
}

pipeline_descriptions = {
    "create_pipeline": "Create a new pipeline with the specified configurations.",
    "get_pipeline": "Retrieve the configuration of a specific pipeline by pipeline ID.",
    "get_all_pipelines": "Retrieve a list of all available pipelines.",
    "update_pipeline": "Update the configuration of a specific pipeline by pipeline ID.",
    "delete_pipeline": "Delete a specific pipeline by pipeline ID.",
    "start_pipeline": "Start a specific pipeline by pipeline ID.",
    "stop_pipeline": "Stop a specific pipeline by pipeline ID."
}