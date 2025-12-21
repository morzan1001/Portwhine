#!/usr/bin/env python3
import os
from datetime import datetime
from typing import Optional, Dict, Tuple, Any, List

from selenium import webdriver
from selenium.webdriver.chrome.service import Service
from selenium.webdriver.chrome.options import Options

from utils.base_worker import BaseWorker
from utils.minio import get_minio_client
from models.job_payload import JobPayload, HttpTarget

class ScreenshotWorker(BaseWorker):
    def __init__(self):
        super().__init__()
        self.s3_client = get_minio_client()
        self.bucket_name = os.getenv("MINIO_BUCKET", "screenshots")

    def take_screenshot(self, url: str) -> str:
        """Takes a screenshot of the given URL and returns the file path"""
        resolution = self.config.get('resolution', '1920x1080')

        options = Options()
        options.add_argument("--headless")
        options.add_argument("--no-sandbox")
        options.add_argument("--disable-dev-shm-usage")
        options.add_argument(f"--window-size={resolution}")
        
        service = Service('/usr/bin/chromedriver')
        driver = webdriver.Chrome(service=service, options=options)
        try:
            driver.get(url)
            screenshot_path = f"/tmp/screenshot_{datetime.now().timestamp()}.png"
            driver.save_screenshot(screenshot_path)
            return screenshot_path
        finally:
            driver.quit()

    def upload_to_minio(self, file_path: str) -> str:
        """Uploads the screenshot to MinIO and returns the object URL"""
        object_name = os.path.basename(file_path)
        # Ensure bucket exists
        if not self.s3_client.bucket_exists(self.bucket_name):
            self.s3_client.make_bucket(self.bucket_name)
            
        self.s3_client.fput_object(self.bucket_name, object_name, file_path)
        return f"{self.bucket_name}/{object_name}"

    def execute(self, payload: JobPayload) -> Tuple[Optional[JobPayload], Dict[str, Any]]:
        results = []
        http_targets = []
        
        for http_target in payload.http:
            url = http_target.url
            if not url:
                continue
                
            if not url.startswith("http"):
                url = f"http://{url}"

            self.logger.info(f"Taking screenshot of: {url}")
            try:
                screenshot_path = self.take_screenshot(url)
                object_url = self.upload_to_minio(screenshot_path)
                
                results.append({
                    "url": url,
                    "screenshot_url": object_url,
                    "category": "screenshot"
                })
                # Pass through the target
                http_targets.append(HttpTarget(url=url, method="GET"))
                
                self.logger.info(f"Screenshot processed for {url}")
            except Exception as e:
                self.logger.error(f"Failed to process screenshot for {url}: {e}")

        return JobPayload(http=http_targets), {"screenshots": results}

if __name__ == "__main__":
    worker = ScreenshotWorker()
    worker.run()
