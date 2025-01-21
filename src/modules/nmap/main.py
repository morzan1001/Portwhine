#!/usr/bin/env python3
import os
import sys
import json
import subprocess
import xmltodict
from datetime import datetime
from typing import Optional, Dict, List, Any
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
        """Parses and normalizes the nmap XML output"""
        try:
            # Parse XML to dict
            parsed = xmltodict.parse(output)
            
            # Normalize script elements
            if 'nmaprun' in parsed and 'host' in parsed['nmaprun']:
                hosts = parsed['nmaprun']['host']
                if not isinstance(hosts, list):
                    hosts = [hosts]
                    
                for host in hosts:
                    if 'ports' in host and 'port' in host['ports']:
                        ports = host['ports']['port']
                        if not isinstance(ports, list):
                            ports = [ports]
                            
                        for port in ports:
                            if 'script' in port:
                                scripts = port['script']
                                if not isinstance(scripts, list):
                                    scripts = [scripts]
                                    
                                for script in scripts:
                                    if 'elem' in script and not isinstance(script['elem'], list):
                                        script['elem'] = [script['elem']]

            return parsed
        except Exception as e:
            logger.error(f"Error parsing nmap output: {e}")
            return {}

    def organize_results_by_service(self, results: Dict) -> Dict:
        """Organizes results by service type"""
        organized = {
            "http": [],
            "ip": []
        }
        
        try:
            hosts = results.get('nmaprun', {}).get('host', [])
            if not isinstance(hosts, list):
                hosts = [hosts]
                
            for host in hosts:
                if not isinstance(host, dict):
                    continue
                    
                ip = host.get('address', {}).get('@addr')
                if not ip:
                    continue

                ports = host.get('ports', {}).get('port', [])
                if not isinstance(ports, list):
                    ports = [ports]
                    
                for port in ports:
                    if not isinstance(port, dict):
                        continue
                        
                    service = port.get('service', {}).get('@name', '')
                    port_id = port.get('@portid')
                    
                    if service and port_id:
                        organized["http"].append({
                            "protocol": service,
                            "ip": ip,
                            "port": port_id
                        })
                        
                organized["ip"].append(ip)
                
            organized["ip"] = list(set(organized["ip"]))
            return organized
            
        except Exception as e:
            logger.error(f"Error organizing results: {e}")
            return {"http": [], "ip": []}

    def notify_handler(self, organized_results: Dict):
        """Notifies the handler with organized results"""
        try:
            endpoint = f"/job/{self.pipeline_id}/{self.worker_id}"
            url = f"http://api:8000/api/v1{endpoint}"
            
            payload = {
                "services": organized_results["http"],
                "ips": organized_results["ip"]
            }
            
            headers = {'Content-Type': 'application/json'}
            response = requests.post(url, json=payload, headers=headers)
            
            if response.status_code == 200:
                logger.info(f"Successfully notified handler for pipeline {self.pipeline_id}")
            else:
                logger.error(f"Failed to notify handler: {response.status_code}")
                
        except Exception as e:
            logger.error(f"Error notifying handler: {e}")

    def save_results(self, organized_results: Dict, results: Dict) -> None:
        """Saves sanitized results to Elasticsearch"""
        try:
            document = {
                "metadata": {
                    "scan_date": datetime.now().isoformat(),
                    "category": "portscan",
                    "tags": list(organized_results.keys())
                },
                "organized_results": organized_results,
                "raw_results": self._sanitize_for_es(results)
            }
            
            self.es.index(index=self.pipeline_id, document=document)
            logger.info(f"Results saved to index {self.pipeline_id}")
            
        except Exception as e:
            logger.error(f"Error saving to Elasticsearch: {e}")

    def _sanitize_for_es(self, data: Any) -> Any:
        """Sanitizes data for Elasticsearch storage"""
        if isinstance(data, dict):
            return {k: self._sanitize_for_es(v) for k, v in data.items()}
        elif isinstance(data, list):
            return [self._sanitize_for_es(v) for v in data]
        elif isinstance(data, (str, int, float, bool)):
            return data
        else:
            return str(data)

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
        logger.info(f"Results: {results}")
        organized_results = scanner.organize_results_by_service(results)
        scanner.save_results(organized_results, results)
        scanner.notify_handler(organized_results)
        logger.info(f"Results saved and sent for {payload.get('ip')}")

if __name__ == "__main__":
    main()