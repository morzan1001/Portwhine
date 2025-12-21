#!/usr/bin/env python3
import os
import sys
import time
import json
import threading
import requests
from http.server import HTTPServer, BaseHTTPRequestHandler
from typing import Optional, Dict, List, Union

from utils.logger import LoggingModule
from models.job_payload import JobPayload, IpTarget
from models.worker_result import WorkerResult
from models.types import NodeStatus
from models.trigger import IPAddressTrigger

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

class IPAddressScanner:
    def __init__(self):
        self.pipeline_id = os.getenv("PIPELINE_ID")
        self.worker_id = os.getenv("WORKER_ID")
        self.run_id = os.getenv("RUN_ID")
        self.instance_name = os.getenv("INSTANCE_NAME")
        self.current_status = NodeStatus.STARTING
        
        if not self.pipeline_id or not self.worker_id:
            logger.error("Missing required environment variables: PIPELINE_ID or WORKER_ID")
            sys.exit(1)

        self.config = self._load_config()
        self.ip_addresses = [str(ip) for ip in self.config.ip_addresses]
        self.repetition = self.config.repetition
        
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

    def _load_config(self) -> IPAddressTrigger:
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
            
            return IPAddressTrigger.model_validate(data)
        except Exception as e:
            logger.error(f"Invalid TRIGGER_CONFIG: {e}")
            sys.exit(1)

    def notify_handler(self, ip_address: str):
        """Notifies the handler with the IP address"""
        try:
            url = f"https://api:8000/api/v1/job/result"
            if self.instance_name:
                url += f"?instance_name={self.instance_name}"

            output_payload = JobPayload(
                ip=[IpTarget(ip=ip_address)]
            )
            
            worker_result = WorkerResult(
                run_id=self.run_id,
                pipeline_id=self.pipeline_id,
                node_id=self.worker_id,
                status=NodeStatus.COMPLETED,
                output_payload=output_payload,
                raw_data={"ip": ip_address}
            )

            headers = {'Content-Type': 'application/json'}
            response = requests.post(url, json=worker_result.model_dump(mode='json'), headers=headers, verify='/usr/local/share/ca-certificates/selfsigned-ca.crt')
            
            if response.status_code == 200:
                logger.info(f"Successfully notified handler for pipeline {self.pipeline_id}")
            else:
                logger.error(f"Failed to notify handler for pipeline {self.pipeline_id}: {response.status_code}")
                logger.debug(f"IP-Address: {ip_address} \nResponse: {response.text}")
        except Exception as e:
            logger.error(f"Error notifying handler for pipeline {self.pipeline_id}: {e}")

    def process_ip_addresses(self):
        """Processes the stored IP addresses and notifies the handler"""
        self.current_status = NodeStatus.RUNNING
        while True:
            for ip_address in self.ip_addresses:
                self.notify_handler(ip_address)
            if self.repetition is None:
                break
            time.sleep(self.repetition)
        self.current_status = NodeStatus.COMPLETED

def main():
    scanner = IPAddressScanner()
    scanner.process_ip_addresses()

if __name__ == "__main__":
    main()
