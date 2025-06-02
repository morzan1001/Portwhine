#!/usr/bin/env python3
import json
import os
import sys
from datetime import datetime, timezone

import requests
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options

from utils.elasticsearch import get_elasticsearch_connection
from utils.minio import get_minio_client
from utils.logger import LoggingModule


# Logger initializing
logger = LoggingModule()

class ScreenshotTaker:
    def __init__(self, pipeline_id: str, worker_id: str):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()
        self.pipeline_id = pipeline_id
        self.worker_id = worker_id
        
        # MinIO client setup
        self.s3_client = get_minio_client()
        self.bucket_name = os.getenv("MINIO_BUCKET", "screenshots")

    def take_screenshot(self, url: str) -> str:
        """Takes a screenshot of the given URL and returns the file path"""
        options = Options()
        options.headless = True
        service = Service('/usr/bin/chromedriver')
        driver = webdriver.Chrome(service=service, options=options)
        driver.get(url)
        screenshot_path = f"/tmp/screenshot_{datetime.now().timestamp()}.png"
        driver.save_screenshot(screenshot_path)
        driver.quit()
        return screenshot_path

    def upload_to_minio(self, file_path: str) -> str:
        """Uploads the screenshot to MinIO and returns the object URL"""
        object_name = os.path.basename(file_path)
        self.s3_client.upload_file(file_path, self.bucket_name, object_name)
        object_url = f"{self.s3_client.meta.endpoint_url}/{self.bucket_name}/{object_name}"
        return object_url

    def save_to_elasticsearch(self, url: str, object_url: str):
        """Saves the screenshot metadata to Elasticsearch"""
        document = {
            "url": url,
            "screenshot_url": object_url,
            "timestamp": datetime.now(timezone.utc),
            "category": "screenshot",
        }
        self.es.index(index="screenshots", document=document)
        logger.info(f"Screenshot metadata for {url} saved successfully.")

    def notify_handler(self, http_payload: str):
        """Notifies the handler with the screenshot URL"""
        try:
            endpoint = f"/job/{self.pipeline_id}/{self.worker_id}"
            url = f"http://api:8000/api/v1{endpoint}"
            payload = {"http": http_payload,}
            response = requests.post(url, json=payload)
            response.raise_for_status()
            logger.info(f"Notification sent to handler for {http_payload}")
        except Exception as e:
            logger.error(f"Error notifying handler: {e}")

def main():
    payload = os.getenv('JOB_PAYLOAD')
    if not payload:
        logger.error("JOB_PAYLOAD environment variable not set")
        sys.exit(1)
    else:
        payload = json.loads(payload)
    
    pipeline_id = os.getenv("PIPELINE_ID")
    if not pipeline_id:
        logger.error("PIPELINE_ID environment variable not set.")
        return
    
    worker_id = os.getenv("WORKER_ID")
    if not worker_id:
        logger.error("WORKER_ID environment variable not set.")
        return
    
    scanner = ScreenshotTaker(pipeline_id, worker_id)
    for http_payload in payload.get("http", []):
        logger.info(f"Taking Screenhot: {http_payload}")
        result = scanner.take_screenshot(http_payload)
        if result:
            object_url = scanner.upload_to_minio(result)
            scanner.save_to_elasticsearch(http_payload, object_url)
            scanner.notify_handler(http_payload)
            logger.info(f"Results saved and sent for {http_payload}")

if __name__ == "__main__":
    main()