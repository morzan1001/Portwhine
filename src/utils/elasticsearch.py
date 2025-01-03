#!/usr/bin/env python3
import os
from typing import Optional
from elasticsearch import Elasticsearch
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule.get_logger()

def get_elasticsearch_connection() -> Optional[Elasticsearch]:
    """Returns an Elasticsearch connection or None if an error occurs"""
    try:
        es = Elasticsearch(
            hosts=[os.getenv("DATABASE_HOST", "http://elasticsearch:9200")],
            basic_auth=(
                os.getenv("DATABASE_USER", "elastic"),
                os.getenv("DATABASE_PASSWORD", "changeme")
            )
        )
        # Test the connection
        if es.ping():
            logger.info("Successfully connected to Elasticsearch")
        else:
            logger.warning("Elasticsearch ping failed")
        return es
    except Exception as e:
        logger.error(f"Error connecting to Elasticsearch: {e}")
        return None