#!/usr/bin/env python3
import os
import json
import subprocess
from typing import Optional, Dict, Tuple, Any, List

from utils.base_worker import BaseWorker
from models.job_payload import JobPayload, HttpTarget

class HumbleWorker(BaseWorker):
    def scan_url(self, url: str) -> Optional[Dict]:
        """Scans a URL using humble directly"""
        try:
            # humble.py path might vary in docker image
            humble_path = '/humble/humble.py'
            if not os.path.exists(humble_path):
                 # Try to find it or assume it's in PATH if installed differently
                 humble_path = 'humble' 

            result = subprocess.run(['python3', humble_path, '-u', url, '-o', 'json'], capture_output=True, text=True)
            if result.returncode == 0:
                try:
                    return json.loads(result.stdout)
                except json.JSONDecodeError:
                    self.logger.error(f"Failed to decode humble output: {result.stdout}")
                    return None
            else:
                self.logger.error(f"Error scanning {url}: {result.stderr}")
                return None
        except Exception as e:
            self.logger.error(f"Error scanning {url}: {e}")
            return None

    def execute(self, payload: JobPayload) -> Tuple[Optional[JobPayload], Dict[str, Any]]:
        results = []
        http_targets = []
        
        for http_target in payload.http:
            url = http_target.url
            if not url:
                continue
                
            if not url.startswith("http"):
                url = f"http://{url}"

            self.logger.info(f"Scanning URL: {url}")
            scan_result = self.scan_url(url)
            if scan_result:
                results.append({
                    "url": url,
                    "scan_result": scan_result
                })
                # Pass through the target
                http_targets.append(HttpTarget(url=url, method="GET"))
                self.logger.info(f"Results processed for {url}")

        return JobPayload(http=http_targets), {"headers": results}

if __name__ == "__main__":
    worker = HumbleWorker()
    worker.run()
