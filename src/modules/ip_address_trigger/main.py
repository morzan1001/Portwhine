#!/usr/bin/env python3
import os
import time
from typing import Optional, Dict, List, Union

import requests

from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule.get_logger()

class IPAddressScanner:
    def __init__(self, pipeline_id: str):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()
        self.pipeline_id = pipeline_id
        self.ip_addresses = self.get_stored_ip_addresses()
        self.job_id = self.get_job_id()
        self.repetition = self.get_repetition_interval()

    def get_stored_ip_addresses(self) -> Union[List[str], List[Dict[str, str]]]:
        """Fetches stored IP addresses from Elasticsearch based on pipeline_id"""
        try:
            doc = self.es.get(index="pipelines", id=self.pipeline_id)
            pipeline = doc["_source"]
            if "trigger" in pipeline and "IPAddressTrigger" in pipeline["trigger"]:
                return pipeline["trigger"]["IPAddressTrigger"]["ip_addresses"]
        except Exception as e:
            logger.error(f"Error fetching IP addresses for pipeline {self.pipeline_id}: {e}")
            return []

    def get_job_id(self) -> str:
        """Fetches the job ID from Elasticsearch based on pipeline_id"""
        try:
            doc = self.es.get(index="pipelines", id=self.pipeline_id)
            pipeline = doc["_source"]
            if "trigger" in pipeline and "IPAddressTrigger" in pipeline["trigger"]:
                return pipeline["trigger"]["IPAddressTrigger"]["id"]
        except Exception as e:
            logger.error(f"Error fetching container ID for pipeline {self.pipeline_id}: {e}")
            return ""

    def get_repetition_interval(self) -> Optional[int]:
        """Fetches the repetition interval from Elasticsearch based on pipeline_id"""
        try:
            doc = self.es.get(index="pipelines", id=self.pipeline_id)
            pipeline = doc["_source"]
            if "trigger" in pipeline and "IPAddressTrigger" in pipeline["trigger"]:
                return pipeline["trigger"]["IPAddressTrigger"].get("repetition")
        except Exception as e:
            logger.error(f"Error fetching repetition interval for pipeline {self.pipeline_id}: {e}")
            return None

    def notify_handler(self, ip_address: str):
        """Notifies the handler with the IP address"""
        try:
            endpoint = f"/job/{self.pipeline_id}/{self.job_id}"
            url = f"https://api:8000/api/v1{endpoint}"
            payload = {"ip": [ip_address]}
            headers = {'Content-Type': 'application/json'}
            response = requests.post(url, json=payload, headers=headers, verify='/usr/local/share/ca-certificates/selfsigned-ca.crt')
            if response.status_code == 200:
                logger.info(f"Successfully notified handler for pipeline {self.pipeline_id}")
            else:
                logger.error(f"Failed to notify handler for pipeline {self.pipeline_id}: {response.status_code}")
                logger.debug(f"IP-Address: {ip_address} \nResponse: {response.text}")
        except Exception as e:
            logger.error(f"Error notifying handler for pipeline {self.pipeline_id}: {e}")

    def process_ip_addresses(self):
        """Processes the stored IP addresses and notifies the handler"""
        while True:
            for ip_address in self.ip_addresses:
                self.notify_handler(ip_address)
            if self.repetition is None:
                break
            time.sleep(self.repetition)

def main():
    pipeline_id = os.getenv("PIPELINE_ID")
    if not pipeline_id:
        logger.error("PIPELINE_ID environment variable not set.")
        return
    scanner = IPAddressScanner(pipeline_id)
    scanner.process_ip_addresses()

if __name__ == "__main__":
    main()