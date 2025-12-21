#!/usr/bin/env python3
import os
from utils.logger import LoggingModule
from typing import Optional
import redis

# Logger initializing
logger = LoggingModule.get_logger()

def get_redis_connection(db: int = 0) -> Optional[redis.Redis]:
    """Returns a Redis connection or None if an error occurs"""
    try:
        redis_client = redis.Redis(
            host=os.getenv("REDIS_HOST", "localhost"),
            port=int(os.getenv("REDIS_PORT", 6379)),
            username=os.getenv("REDIS_USER", None),
            password=os.getenv("REDIS_PASSWORD", None),
            db=db,
            ssl=True,
            ssl_ca_certs=os.getenv("CA_CERT_PATH", "/certs/ca.crt"),
            ssl_certfile=os.getenv("CLIENT_CERT_PATH", "/certs/client.crt"),
            ssl_keyfile=os.getenv("CLIENT_KEY_PATH", "/certs/client.key"),
        )
        # Test the connection
        if redis_client.ping():
            logger.info("Successfully connected to Redis")
        else:
            logger.warning("Redis ping failed")
        return redis_client
    except Exception as e:
        logger.error(f"Error connecting to Redis: {e}")
        return None