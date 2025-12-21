#!/usr/bin/env python3
import os
import socket
from typing import Optional, Dict, Tuple, Any, List
from urllib.parse import urlparse

from utils.base_worker import BaseWorker
from utils.helper import is_private_ip
from models.job_payload import JobPayload, IpTarget

class ResolverWorker(BaseWorker):
    def __init__(self):
        super().__init__()
        self.use_internal = self.config.get('use_internal', False)

    def resolve_domain(self, domain: str) -> Optional[str]:
        """Resolves a domain to an IP address"""
        try:
            ip_address = socket.gethostbyname(domain)
            if not self.use_internal and is_private_ip(ip_address):
                self.logger.error(f"Resolved IP address for {domain} is private: {ip_address}")
                return None
            return ip_address
        except socket.gaierror as e:
            self.logger.error(f"Error resolving {domain}: {e}")
            return None

    def execute(self, payload: JobPayload) -> Tuple[Optional[JobPayload], Dict[str, Any]]:
        results = []
        ip_targets = []
        
        for http_target in payload.http:
            url = http_target.url
            if not url:
                continue
            
            # Extract domain from URL
            try:
                if not url.startswith("http"):
                    url_to_parse = "http://" + url
                else:
                    url_to_parse = url
                parsed = urlparse(url_to_parse)
                domain = parsed.hostname
            except Exception:
                domain = url # Fallback

            if domain:
                self.logger.info(f"Resolving domain: {domain}")
                ip_address = self.resolve_domain(domain)
                if ip_address:
                    results.append({
                        "domain": domain,
                        "ip": ip_address
                    })
                    ip_targets.append(IpTarget(ip=ip_address))
                    self.logger.info(f"Resolved {domain} to {ip_address}")

        return JobPayload(ip=ip_targets), {"resolutions": results}

if __name__ == "__main__":
    worker = ResolverWorker()
    worker.run()
