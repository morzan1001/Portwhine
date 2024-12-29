#!/usr/bin/env python3
import os
import sys
import json
import socket
from datetime import datetime
from typing import Optional
import requests

from utils.elasticsearch import get_elasticsearch_connection
from utils.helper import is_private_ip
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule.get_logger()

class Resolver:
    def __init__(self, pipeline_id: str, worker_id: str):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()
        self.pipeline_id = pipeline_id
        self.worker_id = worker_id

    def resolve_domain(self, domain: str) -> Optional[str]:
        """Resolves a domain to an IP address"""
        try:
            ip_address = socket.gethostbyname(domain)
            if is_private_ip(ip_address):
                logger.error(f"Resolved IP address for {domain} is private: {ip_address}")
                return None
            return ip_address
        except socket.gaierror as e:
            logger.error(f"Error resolving {domain}: {e}")
            return None

    def save_results(self, domain: str, ip_address: str) -> None:
        """Saves resolution results to Elasticsearch"""
        try:
            document = {
                "domain": domain,
                "ip_address": ip_address,
                "resolution_date": datetime.now().isoformat(),
                "category": "resolution",
                "tags": ["ip"]
            }
            self.es.index(index=self.pipeline_id, document=document)
            logger.info(f"Results saved successfully in index {self.pipeline_id}.")
        except Exception as e:
            logger.error(f"Error saving results: {e}")

    def notify_handler(self, domain: str, ip_address: str):
        """Notifies the handler with the resolved IP"""
        try:
            endpoint = f"/job/{self.pipeline_id}/{self.worker_id}"
            url = f"http://api:8000/api/v1{endpoint}"
            payload = {"http": [{"domain": domain}],"ip": [ip_address]}
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
        logger.debug(f"Payload: {payload}")
    
    pipeline_id = os.getenv("PIPELINE_ID")
    if not pipeline_id:
        logger.error("PIPELINE_ID environment variable not set.")
        return
    
    worker_id = os.getenv("WORKER_ID")
    if not worker_id:
        logger.error("WORKER_ID environment variable not set.")
        return

    resolver = Resolver(pipeline_id, worker_id)
    http_entries = payload.get('http', [])

    for entry in http_entries:
        domain = entry.get('domain')
        if domain:
            logger.info(f"Resolving Domain: {domain}")
            ip_address = resolver.resolve_domain(domain)
            if ip_address:
                resolver.save_results(domain, ip_address)
                resolver.notify_handler(domain, ip_address)

if __name__ == "__main__":
    main()