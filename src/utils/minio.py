#!/usr/bin/env python3
import os
from typing import Optional
from logger import LoggingModule
import boto3
from botocore.exceptions import BotoCoreError, NoCredentialsError
from botocore.client import BaseClient

# Logger initializing
logger = LoggingModule.get_logger()

def get_minio_client() -> Optional[BaseClient]:
    """Returns a MinIO client or None if an error occurs"""
    try:
        s3_client = boto3.client(
            's3',
            endpoint_url=os.getenv("MINIO_ENDPOINT", "http://minio:9000"),
            aws_access_key_id=os.getenv("MINIO_ROOT_USER", "minioadmin"),
            aws_secret_access_key=os.getenv("MINIO_ROOT_PASSWORD", "minioadmin"),
        )
        logger.info("MinIO client successfully created")
        return s3_client
    except (BotoCoreError, NoCredentialsError) as e:
        logger.error(f"Failed to create MinIO client: {e}")
        return None
