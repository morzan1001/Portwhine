#!/usr/bin/env python3
import json
import re
import threading
import certstream
import os
import sys
import requests
from http.server import HTTPServer, BaseHTTPRequestHandler
from typing import Optional, Dict, List

from utils.logger import LoggingModule
from models.job_payload import JobPayload, HttpTarget
from models.worker_result import WorkerResult
from models.types import NodeStatus
from models.trigger import CertstreamTrigger

# Logger initializing
logger = LoggingModule.get_logger()

class HealthRequestHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path == '/health':
            self.send_response(200)
            self.send_header('Content-type', 'application/json')
            self.end_headers()
            
            status_data = {
                "status": self.server.worker_instance.current_status,
                "worker_id": self.server.worker_instance.worker_id,
                "instance_name": self.server.worker_instance.instance_name
            }
            self.wfile.write(json.dumps(status_data).encode())
        else:
            self.send_response(404)
            self.end_headers()
    
    def log_message(self, format, *args):
        pass

class CertstreamScanner:
    def __init__(self):
        self.regex_lock = threading.Lock()
        self.pipeline_id = os.getenv("PIPELINE_ID")
        self.worker_id = os.getenv("WORKER_ID")
        self.run_id = os.getenv("RUN_ID")
        self.instance_name = os.getenv("INSTANCE_NAME")
        self.current_status = NodeStatus.STARTING
        
        if not self.pipeline_id or not self.worker_id:
            logger.error("Missing required environment variables: PIPELINE_ID or WORKER_ID")
            sys.exit(1)

        self.config = self._load_config()
        self.regexes = [re.compile(self.config.regex)]
        
        self._start_health_server()

    def _start_health_server(self):
        try:
            server = HTTPServer(('0.0.0.0', 8000), HealthRequestHandler)
            server.worker_instance = self
            thread = threading.Thread(target=server.serve_forever)
            thread.daemon = True
            thread.start()
            logger.info("Health server started on port 8000")
        except Exception as e:
            logger.error(f"Failed to start health server: {e}")

    def _load_config(self) -> CertstreamTrigger:
        config_str = os.getenv('TRIGGER_CONFIG')
        if not config_str:
            logger.error("TRIGGER_CONFIG environment variable not set")
            sys.exit(1)
        
        try:
            import json
            data = json.loads(config_str)
            # Unwrap if wrapped in class name
            if isinstance(data, dict) and len(data) == 1:
                first_value = list(data.values())[0]
                if isinstance(first_value, dict):
                    data = first_value
            
            return CertstreamTrigger.model_validate(data)
        except Exception as e:
            logger.error(f"Invalid TRIGGER_CONFIG: {e}")
            sys.exit(1)

    def notify_handler(self, domain: str, message: Dict):
        """Notifies the handler with the results"""
        try:
            url_endpoint = f"https://api:8000/api/v1/job/result"
            if self.instance_name:
                url_endpoint += f"?instance_name={self.instance_name}"

            output_payload = JobPayload(
                http=[HttpTarget(url=domain, method="GET")]
            )
            
            worker_result = WorkerResult(
                run_id=self.run_id,
                pipeline_id=self.pipeline_id,
                node_id=self.worker_id,
                status=NodeStatus.COMPLETED,
                output_payload=output_payload,
                raw_data=message
            )

            headers = {'Content-Type': 'application/json'}
            response = requests.post(url_endpoint, json=worker_result.model_dump(mode='json'), headers=headers, verify='/usr/local/share/ca-certificates/selfsigned-ca.crt')
            
            if response.status_code == 200:
                logger.info(f"Successfully notified handler for {domain}")
            else:
                logger.error(f"Failed to notify handler: {response.status_code}")
        except Exception as e:
            logger.error(f"Error notifying handler: {e}")

    def analyze_cert(self, message: Dict, context):
        """Analyzes a certificate message using stored regex patterns"""
        try:
            if message["message_type"] == "heartbeat":
                return

            if message["message_type"] == "certificate_update":
                all_domains = message["data"]["leaf_cert"]["all_domains"]
                
                for domain in all_domains:
                    with self.regex_lock:
                        for regex in self.regexes:
                            if regex.search(domain):
                                logger.info(f"Match found: {domain}")
                                self.notify_handler(domain, message)
        except Exception as e:
            logger.error(f"Error analyzing certificate: {e}")

def main():
    scanner = CertstreamScanner()
    scanner.current_status = NodeStatus.RUNNING
    logger.info("Starting Certstream listener...")
    certstream.listen_for_events(scanner.analyze_cert, url='wss://certstream.calidog.io/')

if __name__ == "__main__":
    main()
