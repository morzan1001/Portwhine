#!/usr/bin/env python3
import os
import json
import requests
import re
from typing import Optional, Dict, Tuple, Any, List

from utils.base_worker import BaseWorker
from models.job_payload import JobPayload, HttpTarget

class WebAppAnalyzerWorker(BaseWorker):
    def __init__(self):
        super().__init__()
        self.fingerprints = self.load_fingerprints()
        self.categories = self.load_json('/webappanalyzer/categories.json')
        self.groups = self.load_json('/webappanalyzer/groups.json')

    def load_json(self, filepath: str) -> Dict:
        """Loads a JSON file and returns its content"""
        try:
            with open(filepath, 'r') as file:
                return json.load(file)
        except Exception as e:
            self.logger.error(f"Error loading {filepath}: {e}")
            return {}

    def load_fingerprints(self) -> Dict[str, Dict]:
        """Loads the fingerprints from JSON files"""
        fingerprints = {}
        try:
            if os.path.exists('/webappanalyzer'):
                for filename in os.listdir('/webappanalyzer'):
                    if filename.endswith('.json') and filename != 'categories.json' and filename != 'groups.json':
                        with open(os.path.join('/webappanalyzer', filename), 'r') as file:
                            data = json.load(file)
                            fingerprints.update(data)
            else:
                self.logger.warning("/webappanalyzer directory not found")
        except Exception as e:
            self.logger.error(f"Error loading fingerprints: {e}")
        return fingerprints

    def scan_url(self, url: str) -> Optional[Dict]:
        """Scans a URL by making a web request and extracting relevant information"""
        try:
            response = requests.get(url, timeout=10, verify=False) # verify=False for internal/self-signed
            if response.status_code == 200:
                headers = response.headers
                content = response.text
                
                # Simple analysis logic (simplified from original)
                detected_technologies = []
                
                # Check headers
                for tech, fingerprint in self.fingerprints.items():
                    # This is a simplified check. Real Wappalyzer logic is more complex.
                    # Assuming fingerprint has 'headers', 'html', 'script', 'meta' keys
                    
                    # Check headers
                    if 'headers' in fingerprint:
                        for header_name, header_pattern in fingerprint['headers'].items():
                            if header_name in headers:
                                if re.search(header_pattern, headers[header_name], re.IGNORECASE):
                                    detected_technologies.append(tech)
                                    
                    # Check HTML
                    if 'html' in fingerprint:
                        if isinstance(fingerprint['html'], str):
                             if re.search(fingerprint['html'], content, re.IGNORECASE):
                                 detected_technologies.append(tech)
                        elif isinstance(fingerprint['html'], list):
                            for pattern in fingerprint['html']:
                                if re.search(pattern, content, re.IGNORECASE):
                                    detected_technologies.append(tech)

                return {
                    "technologies": list(set(detected_technologies)),
                    "headers": dict(headers)
                }
            else:
                self.logger.warning(f"Failed to fetch {url}: {response.status_code}")
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

            self.logger.info(f"Analyzing URL: {url}")
            scan_result = self.scan_url(url)
            if scan_result:
                results.append({
                    "url": url,
                    "analysis": scan_result
                })
                # Pass through the target
                http_targets.append(HttpTarget(url=url, method="GET"))
                self.logger.info(f"Analysis processed for {url}")

        return JobPayload(http=http_targets), {"webapp_analysis": results}

if __name__ == "__main__":
    worker = WebAppAnalyzerWorker()
    worker.run()
