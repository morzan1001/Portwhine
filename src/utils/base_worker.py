#!/usr/bin/env python3
import os
import sys
import json
import threading
import requests
from http.server import HTTPServer, BaseHTTPRequestHandler
from typing import Optional, Dict, Tuple, Any
from abc import ABC, abstractmethod

from utils.logger import LoggingModule
from models.job_payload import JobPayload
from models.worker_result import WorkerResult
from models.types import NodeStatus

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
        # Suppress default logging to avoid clutter
        pass

class BaseWorker(ABC):
    def __init__(self):
        self.logger = LoggingModule.get_logger()
        self.pipeline_id = os.getenv("PIPELINE_ID")
        self.worker_id = os.getenv("WORKER_ID")
        self.run_id = os.getenv("RUN_ID")
        self.instance_name = os.getenv("INSTANCE_NAME")
        self.current_status = NodeStatus.STARTING
        
        if not self.pipeline_id or not self.worker_id or not self.run_id:
            self.logger.error("Missing required environment variables: PIPELINE_ID, WORKER_ID, or RUN_ID")
            sys.exit(1)

        self.payload = self._load_payload()
        self.config = self._load_config()
        
        self._start_health_server()

    def _start_health_server(self):
        try:
            server = HTTPServer(('0.0.0.0', 8000), HealthRequestHandler)
            server.worker_instance = self
            thread = threading.Thread(target=server.serve_forever)
            thread.daemon = True
            thread.start()
            self.logger.info("Health server started on port 8000")
        except Exception as e:
            self.logger.error(f"Failed to start health server: {e}")

    def _load_payload(self) -> JobPayload:
        payload_str = os.getenv('JOB_PAYLOAD')
        if not payload_str:
            self.logger.error("JOB_PAYLOAD environment variable not set")
            sys.exit(1)
        
        try:
            return JobPayload.model_validate_json(payload_str)
        except Exception as e:
            self.logger.error(f"Invalid JOB_PAYLOAD: {e}")
            sys.exit(1)

    def _load_config(self) -> Dict[str, Any]:
        """Loads the worker configuration from environment variable"""
        config_str = os.getenv('WORKER_CONFIG')
        if not config_str:
            return {}
        
        try:
            import json
            data = json.loads(config_str)
            # Unwrap if wrapped in class name (which is the default serialization behavior)
            if isinstance(data, dict) and len(data) == 1:
                first_value = list(data.values())[0]
                if isinstance(first_value, dict):
                    return first_value
            return data
        except Exception as e:
            self.logger.warning(f"Invalid WORKER_CONFIG: {e}")
            return {}

    @abstractmethod
    def execute(self, payload: JobPayload) -> Tuple[Optional[JobPayload], Dict[str, Any]]:
        """
        Execute the worker logic.
        Returns a tuple of (output_payload, raw_data).
        output_payload: The JobPayload to pass to the next node.
        raw_data: The raw results of the worker execution.
        """
        pass

    def notify_handler(self, status: NodeStatus, output_payload: Optional[JobPayload], raw_data: Optional[Dict[str, Any]], error: Optional[str] = None):
        """Notifies the handler with the results"""
        try:
            url = f"https://api:8000/api/v1/job/result"
            if self.instance_name:
                url += f"?instance_name={self.instance_name}"

            worker_result = WorkerResult(
                run_id=self.run_id,
                pipeline_id=self.pipeline_id,
                node_id=self.worker_id,
                status=status,
                output_payload=output_payload,
                raw_data=raw_data,
                error=error
            )

            headers = {'Content-Type': 'application/json'}
            # Verify with the CA cert we copied in Dockerfile.base
            response = requests.post(
                url, 
                json=worker_result.model_dump(mode='json'), 
                headers=headers, 
                verify='/usr/local/share/ca-certificates/selfsigned-ca.crt'
            )
            
            if response.status_code == 200:
                self.logger.info(f"Successfully notified handler for pipeline {self.pipeline_id}")
            else:
                self.logger.error(f"Failed to notify handler: {response.status_code}")
                self.logger.error(f"Response: {response.text}")
                
        except Exception as e:
            self.logger.error(f"Error notifying handler: {e}")

    def run(self):
        """Main execution method"""
        try:
            self.current_status = NodeStatus.RUNNING
            output_payload, raw_data = self.execute(self.payload)
            self.current_status = NodeStatus.COMPLETED
            self.notify_handler(NodeStatus.COMPLETED, output_payload, raw_data)
        except Exception as e:
            self.current_status = NodeStatus.ERROR
            self.logger.error(f"Worker execution failed: {e}")
            self.notify_handler(NodeStatus.ERROR, None, None, str(e))
