#!/usr/bin/env python3
import sys
import json
import subprocess
from datetime import datetime
from typing import Optional, Dict

from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule()

class FfufScanner:
    def __init__(self):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()

    def scan_url(self, url: str) -> Optional[Dict]:
        """Scans a URL using ffuf directly"""
        try:
            result = subprocess.run(['ffuf', '-u', f'{url}/FUZZ', '-w', '/path/to/wordlist'], capture_output=True, text=True)
            if result.returncode == 0:
                return json.loads(result.stdout)
            else:
                logger.error(f"Error scanning {url}: {result.stderr}")
                return None
        except Exception as e:
            logger.error(f"Error scanning {url}: {e}")
            return None

    def save_results(self, url: str, results: Dict) -> None:
        """Saves scan results to Elasticsearch"""
        try:
            doc = {
                "url": url,
                "scan_results": results,
                "scan_date": datetime.now().isoformat()
            }
            self.es.index(index="urls", document=doc)
            logger.info(f"Results saved for {url}")
        except Exception as e:
            logger.error(f"Error saving results: {e}")

def main():
    if len(sys.argv) != 2:
        logger.error("Usage: python main.py <url>")
        sys.exit(1)

    url = sys.argv[1]
    scanner = FfufScanner()
    logger.info(f"Scanning URL: {url}")
    results = scanner.scan_url(url)
    if results:
        scanner.save_results(url, results)
        logger.info(f"Results saved for {url}")

if __name__ == "__main__":
    main()