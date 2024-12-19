#!/usr/bin/env python3
import os
import sys
import json
import requests
import re
from datetime import datetime, timezone
from typing import Optional, Dict

from utils.elasticsearch import get_elasticsearch_connection
from utils.logger import LoggingModule

# Logger initializing
logger = LoggingModule()

class WebappScanner:
    def __init__(self):
        # Elasticsearch connection
        self.es = get_elasticsearch_connection()
        self.fingerprints = self.load_fingerprints()
        self.categories = self.load_json('/webappanalyzer/categories.json')
        self.groups = self.load_json('/webappanalyzer/groups.json')

    def load_json(self, filepath: str) -> Dict:
        """Loads a JSON file and returns its content"""
        with open(filepath, 'r') as file:
            return json.load(file)

    def load_fingerprints(self) -> Dict[str, Dict]:
        """Loads the fingerprints from JSON files"""
        fingerprints = {}
        for filename in os.listdir('/webappanalyzer'):
            if filename.endswith('.json') and filename != 'categories.json' and filename != 'groups.json':
                with open(os.path.join('/webappanalyzer', filename), 'r') as file:
                    data = json.load(file)
                    fingerprints.update(data)
        return fingerprints

    def scan_url(self, url: str) -> Optional[Dict]:
        """Scans a URL by making a web request and extracting relevant information"""
        try:
            response = requests.get(url)
            if response.status_code == 200:
                headers = response.headers
                content = response.text
                scripts = re.findall(r'<script[^>]+src="([^"]+)"', content)
                js_vars = re.findall(r'var\s+(\w+)\s*=\s*["\']?([^"\']+)["\']?', content)
                dom_elements = re.findall(r'<([a-z]+)[^>]*>', content)
                return {
                    "headers": headers,
                    "content": content,
                    "scripts": scripts,
                    "js": dict(js_vars),
                    "dom": dom_elements
                }
            else:
                logger.error(f"Error scanning {url}: Status code {response.status_code}")
                return None
        except Exception as e:
            logger.error(f"Error scanning {url}: {e}")
            return None

    def analyze_web_technology(self, url: str) -> Optional[Dict]:
        """Analyzes the web technology of a URL using fingerprints"""
        try:
            response = self.scan_url(url)
            if not response:
                return None

            technologies = []
            for tech_name, details in self.fingerprints.items():
                if self.match_fingerprint(response, details):
                    categories = [self.categories[str(cat_id)]['name'] for cat_id in details.get('cats', [])]
                    groups = [self.groups[str(group_id)]['name'] for cat_id in details.get('cats', []) for group_id in self.categories[str(cat_id)].get('groups', [])]
                    icon_path = os.path.join('images/icons', details.get("icon", ""))
                    tech_info = {
                        "name": tech_name,
                        "categories": categories,
                        "groups": groups,
                        "description": details.get("description", ""),
                        "website": details.get("website", ""),
                        "icon": icon_path,
                        "pricing": details.get("pricing", []),
                        "saas": details.get("saas", False),
                        "implies": details.get("implies", []),
                        "requires": details.get("requires", []),
                        "oss": details.get("oss", False)
                    }
                    technologies.append(tech_info)

            if technologies:
                return {"url": url, "technologies": technologies}
            else:
                logger.error(f"Error analyzing {url}: No matching technologies found")
                return None
        except Exception as e:
            logger.error(f"Error analyzing {url}: {e}")
            return None

    def match_fingerprint(self, response: Dict, details: Dict) -> bool:
        """Matches the response with the given fingerprint details"""
        if 'headers' in details:
            for header, pattern in details['headers'].items():
                if header in response['headers'] and re.search(pattern, response['headers'][header]):
                    return True
        if 'scriptSrc' in details:
            for script in details['scriptSrc']:
                if any(re.search(script, src) for src in response.get('scripts', [])):
                    return True
        if 'js' in details:
            for js_var, pattern in details['js'].items():
                if js_var in response.get('js', {}) and re.search(pattern, response['js'][js_var]):
                    return True
        if 'dom' in details:
            for dom_selector in details['dom']:
                if any(re.search(dom_selector, element) for element in response.get('dom', [])):
                    return True
        if 'meta' in details:
            for meta_name, pattern in details['meta'].items():
                if meta_name in response.get('meta', {}) and re.search(pattern, response['meta'][meta_name]):
                    return True
        if 'html' in details:
            for html_pattern in details['html']:
                if re.search(html_pattern, response.get('content', '')):
                    return True
        return False

    def save_results(self, url: str, results: Dict):
        """Saves the scan results to Elasticsearch"""
        try:
            document = {
                "url": url,
                "results": results,
                "timestamp": datetime.now(timezone.utc)
            }
            self.es.index(index="webscan", document=document)
            logger.info(f"Results for {url} saved successfully.")
        except Exception as e:
            logger.error(f"Error saving results for {url}: {e}")

def main():
    if len(sys.argv) != 2:
        logger.error("Usage: python main.py <url>")
        sys.exit(1)

    url = sys.argv[1]
    scanner = WebappScanner()
    scan_results = scanner.scan_url(url)
    if scan_results:
        scanner.save_results(url, scan_results)
    tech_results = scanner.analyze_web_technology(url)
    if tech_results:
        scanner.save_results(url, tech_results)

if __name__ == "__main__":
    main()