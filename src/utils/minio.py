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
            endpoint_url=os.getenv("MINIO_ENDPOINT", "https://minio:9000"),
            aws_access_key_id=os.getenv("APP_MINIO_USER"),
            aws_secret_access_key=os.getenv("APP_MINIO_PASSWORD"),
            verify=os.getenv("CA_CERT_PATH", "/certs/ca.crt")
        )
        logger.info("MinIO client successfully created")
        return s3_client
    except (BotoCoreError, NoCredentialsError) as e:
        logger.error(f"Failed to create MinIO client: {e}")
        return None
