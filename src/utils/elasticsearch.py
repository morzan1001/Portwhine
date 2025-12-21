#!/usr/bin/env python3
import os
from typing import Optional
from elasticsearch import Elasticsearch
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule.get_logger()

def get_elasticsearch_connection() -> Optional[Elasticsearch]: # type: ignore
    """Returns an Elasticsearch connection or None if an error occurs"""
    try:
        es = Elasticsearch(
            hosts=[os.getenv("DATABASE_HOST", "https://elasticsearch:9200")],
            basic_auth=(
                os.getenv("DATABASE_USER", "elastic"),
                os.getenv("DATABASE_PASSWORD", "changeme")
            ),
            ca_certs=os.getenv("CA_CERT_PATH", "/certs/ca.crt"),
            client_cert=os.getenv("CLIENT_CERT_PATH", "/certs/client.crt"),
            client_key=os.getenv("CLIENT_KEY_PATH", "/certs/client.key"),
            verify_certs=True
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