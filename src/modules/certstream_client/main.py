#!/usr/bin/env python3
import re
import threading
import certstream
from datetime import datetime, timezone
from typing import Optional, Dict

from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule()

class CertstreamScanner:
    def __init__(self):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()
        self.regex_lock = threading.Lock()
        self.regexes = self.get_stored_regexes()

    def get_stored_regexes(self) -> list:
        """Fetches stored regex patterns from Elasticsearch"""
        try:
            doc = self.es.get(index="regex-store", id="1")
            return [re.compile(regex["pattern"]) for regex in doc["_source"].get("regex_list", [])]
        except Exception as e:
            logger.error(f"Error fetching regexes: {e}")
            return []

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
                return {"cert_data": cert_data, "matched_patterns": matched_patterns}
            else:
                logger.info(f"No patterns matched for certificate: {cert_data}")
                return None
        except Exception as e:
            logger.error(f"Error analyzing certificate: {e}")
            return None

    def save_results(self, results: Dict):
        """Saves the analysis results to Elasticsearch"""
        try:
            document = {
                "results": results,
                "timestamp": datetime.now(timezone.utc)
            }
            self.es.index(index="certstream", document=document)
            logger.info("Results saved successfully.")
        except Exception as e:
            logger.error(f"Error saving results: {e}")

    def certstream_callback(self, message: Dict):
        """Callback function for certstream messages"""
        if message['message_type'] == "certificate_update":
            results = self.analyze_cert(message)
            if results:
                self.save_results(results)

def main():
    scanner = CertstreamScanner()
    certstream.listen_for_events(scanner.certstream_callback)

if __name__ == "__main__":
    main()