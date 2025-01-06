#!/usr/bin/env python3
import json
import re
import threading
import certstream
from datetime import datetime, timezone
from typing import Optional, Dict
import requests
import os

from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule.get_logger()

class CertstreamScanner:
    def __init__(self, pipeline_id: str):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()
        self.regex_lock = threading.Lock()
        self.pipeline_id = pipeline_id
        self.regexes = self.get_stored_regexes()
        self.job_id = self.get_job_id()

    def get_stored_regexes(self) -> list:
        """Fetches stored regex patterns from Elasticsearch based on pipeline_id"""
        try:
            doc = self.es.get(index="pipelines", id=self.pipeline_id)
            pipeline = doc["_source"]
            if "trigger" in pipeline and "CertstreamTrigger" in pipeline["trigger"]:
                return [re.compile(pipeline["trigger"]["CertstreamTrigger"]["regex"])]
        except Exception as e:
            logger.error(f"Error fetching regexes for pipeline {self.pipeline_id}: {e}")
            return []
        
    def get_job_id(self) -> str:
        """Fetches the job ID from Elasticsearch based on pipeline_id"""
        try:
            doc = self.es.get(index="pipelines", id=self.pipeline_id)
            pipeline = doc["_source"]
            if "trigger" in pipeline and "CertstreamTrigger" in pipeline["trigger"]:
                return pipeline["trigger"]["CertstreamTrigger"]["id"]
        except Exception as e:
            logger.error(f"Error fetching container ID for pipeline {self.pipeline_id}: {e}")
            return ""

    def analyze_cert(self, message: Dict) -> Optional[Dict]:
        """Analyzes a certificate message using stored regex patterns"""
        try:
            cert_data = message['data']['leaf_cert']['all_domains']
            matched_patterns = []
            for domain in cert_data:
                for regex in self.regexes:
                    if regex.search(domain):
                        matched_patterns.append(domain)
            if matched_patterns:
                return {"cert_data": message, "urls": matched_patterns}
            else:
                logger.debug(f"No patterns matched for certificate: {cert_data}")
                return None
        except Exception as e:
            logger.error(f"Error analyzing certificate: {e}")
            return None
        
    def save_results(self, results: Dict):
        """Saves the analysis results to Elasticsearch"""
        try:
            if "cert_data" in results:
                results["cert_data"] = json.dumps(results["cert_data"])

            document = {
                "results": results,
                "timestamp": datetime.now(timezone.utc),
                "category": "certificate",
                "tags": ["http"]
            }
            self.es.index(index=self.pipeline_id, document=document)
            logger.info(f"Results saved successfully in index {self.pipeline_id}.")
        except Exception as e:
            logger.error(f"Error saving results: {e}")
            logger.debug(f"Results data: {results}")

    def notify_handler(self, matched_url: str):
        """Notifies the handler with the matched URL"""
        try:
            endpoint = f"/job/{self.pipeline_id}/{self.job_id}"
            url = f"https://api:8000/api/v1{endpoint}"
            payload = {"http": [{"domain": matched_url}]}
            headers = {'Content-Type': 'application/json'}
            response = requests.post(url, json=payload, headers=headers, verify='/usr/local/share/ca-certificates/selfsigned-ca.crt')
            print(response.status_code)
            print(response.text)
            if response.status_code == 200:
                logger.info(f"Successfully notified handler for pipeline {self.pipeline_id}")
            else:
                logger.error(f"Failed to notify handler for pipeline {self.pipeline_id}: {response.status_code}")
        except Exception as e:
            logger.error(f"Error notifying handler for pipeline {self.pipeline_id}: {e}")

    def certstream_callback(self, message: Dict, context: Optional[Dict] = None):
        """Callback function for certstream_trigger messages"""
        if message['message_type'] == "certificate_update":
            results = self.analyze_cert(message)
            if results:
                self.save_results(results)
                for url in results["urls"]:
                    self.notify_handler(url)

def main():
    pipeline_id = os.getenv("PIPELINE_ID")
    if not pipeline_id:
        logger.error("PIPELINE_ID environment variable not set.")
        return
    scanner = CertstreamScanner(pipeline_id)
    certstream.listen_for_events(scanner.certstream_callback, url='wss://certstream.calidog.io/')

if __name__ == "__main__":
    main()