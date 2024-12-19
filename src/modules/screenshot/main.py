#!/usr/bin/env python3
import os
import sys
from datetime import datetime, timezone
from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options

from utils.elasticsearch import get_elasticsearch_connection
from utils.minio import get_minio_client
from utils.logger import LoggingModule


# Logger initializing
logger = LoggingModule()

class ScreenshotTaker:
    def __init__(self):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()
        
        # MinIO client setup
        self.s3_client = get_minio_client()
        self.bucket_name = os.getenv("MINIO_BUCKET", "screenshots")

    def take_screenshot(self, url: str) -> str:
        """Takes a screenshot of the given URL and returns the file path"""
        options = Options()
        options.headless = True
        service = Service('/path/to/chromedriver')
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
            "timestamp": datetime.now(timezone.utc)
        }
        self.es.index(index="screenshots", document=document)
        logger.info(f"Screenshot metadata for {url} saved successfully.")

def main():
    if len(sys.argv) != 2:
        logger.error("Usage: python main.py <url>")
        sys.exit(1)

    url = sys.argv[1]
    screenshot_taker = ScreenshotTaker()
    screenshot_path = screenshot_taker.take_screenshot(url)
    object_url = screenshot_taker.upload_to_minio(screenshot_path)
    screenshot_taker.save_to_elasticsearch(url, object_url)

if __name__ == "__main__":
    main()