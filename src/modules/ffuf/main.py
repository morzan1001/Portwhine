#!/usr/bin/env python3
import os
import json
import subprocess
from typing import Optional, Dict, Tuple, Any, List

from utils.base_worker import BaseWorker
from models.job_payload import JobPayload, HttpTarget

class FfufWorker(BaseWorker):
    def scan_url(self, url: str) -> Optional[Dict]:
        """Scans a URL using ffuf directly"""
        try:
            wordlist = self.config.get('wordlist', "/usr/share/wordlists/common.txt")
            extensions = self.config.get('extensions')
            recursive = self.config.get('recursive', False)

            if not os.path.exists(wordlist):
                 self.logger.warning(f"Wordlist not found at {wordlist}")
                 return None

            # Construct command
            cmd = ['ffuf', '-u', f'{url}/FUZZ', '-w', wordlist, '-o', '-', '-of', 'json']
            
            if extensions:
                cmd.extend(['-e', extensions])
            
            if recursive:
                cmd.append('-recursion')

            # -o - -of json outputs to stdout in json format
            result = subprocess.run(cmd, capture_output=True, text=True)
            if result.returncode == 0 or result.returncode == 1: # ffuf returns 0 on success, sometimes 1 if matches found?
                try:
                    return json.loads(result.stdout)
                except json.JSONDecodeError:
                    self.logger.error(f"Failed to decode ffuf output: {result.stdout}")
                    return None
            else:
                self.logger.error(f"Error scanning {url}: {result.stderr}")
                return None
        except Exception as e:
            self.logger.error(f"Error scanning {url}: {e}")
            return None

    def execute(self, payload: JobPayload) -> Tuple[Optional[JobPayload], Dict[str, Any]]:
        all_results = []
        all_http_targets = []
        
        for http_target in payload.http:
            url = http_target.url
            if not url:
                continue
                
            if not url.startswith("http"):
                url = f"http://{url}"

            self.logger.info(f"Scanning URL: {url}")
            results = self.scan_url(url)
            if results:
                all_results.append({"url": url, "scan_data": results})
                
                # Parse ffuf results to find discovered paths
                if 'results' in results:
                    for item in results['results']:
                        discovered_url = item.get('url')
                        if discovered_url:
                            all_http_targets.append(HttpTarget(url=discovered_url, method="GET"))
                            
        return JobPayload(http=all_http_targets), {"scans": all_results}

if __name__ == "__main__":
    worker = FfufWorker()
    worker.run()
