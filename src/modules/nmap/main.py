#!/usr/bin/env python3
import os
import sys
import json
import subprocess
import xmltodict
from datetime import datetime
from typing import Optional, Dict, List
import requests

from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule.get_logger()

class NmapScanner:
    def __init__(self, pipeline_id: str, worker_id: str):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()
        self.pipeline_id = pipeline_id
        self.worker_id = worker_id

    def scan_ip(self, targets: str) -> Optional[Dict]:
        """Scans an IP using nmap directly"""
        try:
            target_str = ' '.join(targets)
            result = subprocess.run(['nmap', '-A', '-p-', '-oX', '-', target_str], capture_output=True, text=True)
            if result.returncode == 0:
                return self.parse_nmap_output(result.stdout)
            else:
                logger.error(f"Error scanning {targets}: {result.stderr}")
                return None
        except Exception as e:
            logger.error(f"Error scanning {targets}: {e}")
            return None

    def parse_nmap_output(self, output: str) -> Dict:
        """Parses the nmap XML output into a structured dictionary"""
        try:
            parsed_output = xmltodict.parse(output)
            return json.loads(json.dumps(parsed_output))
        except Exception as e:
            logger.error(f"Error parsing nmap output: {e}")
            return {}

    def save_results(self, organized_results: Dict ,results: Dict) -> None:
        """Saves scan results to Elasticsearch"""
        tags = list(organized_results.keys())

        try:
            # Ensure all fields are correctly formatted to avoid type conflicts
            def sanitize_data(data):
                if isinstance(data, dict):
                    return {k: sanitize_data(v) for k, v in data.items()}
                elif isinstance(data, list):
                    return [sanitize_data(v) for v in data]
                elif isinstance(data, (str, int, float, bool)):
                    return data
                else:
                    return str(data)

            sanitized_results = sanitize_data(results)

            document = {
                "scan_results": sanitized_results,
                "scan_date": datetime.now().isoformat(),
                "category": "portscan",
                "tags": tags
            }
            self.es.index(index=self.pipeline_id, document=document)
            logger.info(f"Results saved successfully in index {self.pipeline_id}.")
        except Exception as e:
            logger.error(f"Error saving results: {e}")

    def organize_results_by_service(self, results: Dict) -> Dict[str, List[str]]:
        """Organizes results by service"""
        organized_results = {}
        try:
            for host in results.get('nmaprun', {}).get('host', []):
                ip_address = host.get('address', {}).get('@addr')
                for port in host.get('ports', {}).get('port', []):
                    protocol = port.get('service', {}).get('@name')
                    port_id = port.get('@portid')
                    if protocol and port_id:
                        if protocol not in organized_results:
                            organized_results[protocol] = {"ip": ip_address, "ports": []}
                        organized_results[protocol]["ports"].append(port_id)
        except Exception as e:
            logger.error(f"Error organizing results: {e}")
        return organized_results

    def notify_handler(self, organized_results: Dict):
        """Notifies the handler with the matched URL"""
        try:
            endpoint = f"/job/{self.pipeline_id}/{self.worker_id}"
            url = f"http://api:8000/api/v1{endpoint}"
            payload = {protocol: {"ip": data["ip"], "ports": data["ports"]} for protocol, data in organized_results.items()}
            headers = {'Content-Type': 'application/json'}
            response = requests.post(url, json=payload, headers=headers)
            if response.status_code == 200:
                logger.info(f"Successfully notified handler for pipeline {self.pipeline_id}")
            else:
                logger.error(f"Failed to notify handler for pipeline {self.pipeline_id}: {response.status_code}")
        except Exception as e:
            logger.error(f"Error notifying handler for pipeline {self.pipeline_id}: {e}")

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

    scanner = NmapScanner(pipeline_id, worker_id)
    logger.info(f"Scanning Target: {payload.get('ip')}")
    results = scanner.scan_ip(payload.get('ip'))
    if results:
        organized_results = scanner.organize_results_by_service(results)
        scanner.save_results(organized_results, results)
        scanner.notify_handler(organized_results)
        logger.info(f"Results saved and sent for {payload.get('ip')}")

if __name__ == "__main__":
    main()