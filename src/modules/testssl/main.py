#!/usr/bin/env python3
import subprocess
import json
from typing import Optional, Dict, Tuple, Any, List

from utils.base_worker import BaseWorker
from models.job_payload import JobPayload, HttpTarget

class TestSSLWorker(BaseWorker):
    def scan_url(self, url: str) -> Optional[Dict]:
        """Scans a URL using testssl.sh"""
        try:
            # --jsonfile - outputs to stdout if filename is -
            result = subprocess.run(['bash', '/testssl/testssl.sh', '--jsonfile', '-', url], capture_output=True, text=True)
            if result.returncode == 0:
                try:
                    return json.loads(result.stdout)
                except json.JSONDecodeError:
                    # testssl might output other things before json?
                    # Usually it's clean with --jsonfile -
                    # But let's try to find the json part if it fails
                    self.logger.warning(f"Could not decode JSON directly. Output: {result.stdout[:100]}...")
                    return None
            else:
                self.logger.error(f"Error scanning {url} with testssl.sh: {result.stderr}")
                return None
        except Exception as e:
            self.logger.error(f"Error scanning {url} with testssl.sh: {e}")
            return None

    def execute(self, payload: JobPayload) -> Tuple[Optional[JobPayload], Dict[str, Any]]:
        results = []
        http_targets = []
        
        for http_target in payload.http:
            url = http_target.url
            if not url:
                continue
                
            if not url.startswith("http"):
                url = f"https://{url}" # testssl usually targets https

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

        return JobPayload(http=http_targets), {"ssl_scans": results}

if __name__ == "__main__":
    worker = TestSSLWorker()
    worker.run()
