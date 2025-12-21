#!/usr/bin/env python3
import subprocess
import xmltodict
import shlex
from typing import Optional, Dict, List, Tuple, Any

from utils.base_worker import BaseWorker
from models.job_payload import JobPayload, HttpTarget, IpTarget

class NmapWorker(BaseWorker):
    FORBIDDEN_FLAGS = [
        '--interactive', '--resume', '--stylesheet', 
        '-oN', '-oG', '-oS', '-oA', '-iL'
    ]

    def _validate_custom_command(self, cmd: List[str]) -> bool:
        """Validates the custom command for security risks"""
        if not cmd or cmd[0] != 'nmap':
            self.logger.error("Custom command must start with 'nmap'")
            return False
            
        for arg in cmd:
            if arg in self.FORBIDDEN_FLAGS:
                self.logger.error(f"Forbidden flag detected: {arg}")
                return False
            # Check for file output flags with attached values (e.g. -oNfile)
            for flag in self.FORBIDDEN_FLAGS:
                if arg.startswith(flag) and len(arg) > len(flag):
                     self.logger.error(f"Forbidden flag detected: {arg}")
                     return False
        
        # Check for -oX with value other than -
        try:
            if '-oX' in cmd:
                idx = cmd.index('-oX')
                if idx + 1 < len(cmd) and cmd[idx+1] != '-':
                    self.logger.error("Custom command must use '-oX -' for output")
                    return False
        except ValueError:
            pass
            
        return True

    def scan_ip(self, targets: List[str]) -> Optional[Dict]:
        """Scans an IP using nmap directly"""
        try:
            # Sanitize targets
            safe_targets = []
            for t in targets:
                if t.startswith('-'):
                    self.logger.warning(f"Skipping suspicious target: {t}")
                    continue
                safe_targets.append(t)
            
            if not safe_targets:
                self.logger.warning("No valid targets to scan")
                return None

            target_str = ' '.join(safe_targets)
            
            # Get config
            custom_command = self.config.get('custom_command')
            
            if custom_command:
                # Use custom command
                cmd_str = custom_command.replace("{{target}}", target_str)
                cmd = shlex.split(cmd_str)
                
                # Ensure nmap is the command
                if not cmd or cmd[0] != 'nmap':
                    self.logger.error("Custom command must start with 'nmap'")
                    return None

                # Validate
                if not self._validate_custom_command(cmd):
                    return None

                # Enforce -oX - if not present
                if '-oX' not in cmd:
                    cmd.extend(['-oX', '-'])
            else:
                # Standard construction
                ports = self.config.get('ports', '-p-')
                arguments = self.config.get('arguments', '-A')
                
                cmd = ['nmap']
                if arguments:
                    cmd.extend(arguments.split())
                if ports:
                    cmd.append(ports)
                cmd.extend(['-oX', '-', target_str])
            
            self.logger.info(f"Running nmap command: {' '.join(cmd)}")
            
            result = subprocess.run(cmd, capture_output=True, text=True)
            if result.returncode == 0:
                return self.parse_nmap_output(result.stdout)
            else:
                self.logger.error(f"Error scanning {targets}: {result.stderr}")
                return None
        except Exception as e:
            self.logger.error(f"Error scanning {targets}: {e}")
            return None

    def parse_nmap_output(self, output: str) -> Dict:
        """Parses and normalizes the nmap XML output"""
        try:
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
            self.logger.error(f"Error parsing nmap output: {e}")
            return {}

    def organize_results_by_service(self, results: Dict) -> Tuple[List[HttpTarget], List[IpTarget]]:
        """Organizes results by service type"""
        http_targets = []
        ip_targets = []
        
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
                        # Check for http/https services
                        if 'http' in service:
                            protocol = service if service in ['http', 'https'] else 'http'
                            target_url = f"{protocol}://{ip}:{port_id}"
                            http_targets.append(HttpTarget(url=target_url, method="GET"))
                        
                ip_targets.append(IpTarget(ip=ip))
                
            # Deduplicate IPs
            # ip_targets = list(set(ip_targets)) # IpTarget is not hashable by default unless frozen=True
            # Let's just keep them as is or dedupe by string
            
            return http_targets, ip_targets
            
        except Exception as e:
            self.logger.error(f"Error organizing results: {e}")
            return [], []

    def execute(self, payload: JobPayload) -> Tuple[Optional[JobPayload], Dict[str, Any]]:
        targets = [str(ip_target.ip) for ip_target in payload.ip]
        if not targets:
            self.logger.info("No IP targets found in payload.")
            return None, {}

        self.logger.info(f"Scanning Targets: {targets}")
        results = self.scan_ip(targets)
        
        if results:
            http_targets, ip_targets = self.organize_results_by_service(results)
            output_payload = JobPayload(http=http_targets, ip=ip_targets)
            return output_payload, results
        else:
            raise Exception("Scan failed or returned no results.")

if __name__ == "__main__":
    worker = NmapWorker()
    worker.run()
