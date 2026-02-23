package operator

import portwhinev1 "github.com/portwhine/portwhine/gen/go/portwhine/v1"

// getNodeCatalog returns the static catalog of all available node types.
// Each entry mirrors the GetCapabilities response of the corresponding
// worker or trigger container so the frontend can render them without
// starting any containers.
func getNodeCatalog() []*portwhinev1.NodeCatalogEntry {
	return []*portwhinev1.NodeCatalogEntry{
		// ── Triggers ────────────────────────────────────────────────
		{
			Id:          "certstream-trigger",
			DisplayName: "Certstream Monitor",
			Description: "Streams domains from Certificate Transparency logs, with optional domain/regex filtering.",
			NodeType:    "trigger",
			Category:    "monitoring",
			Image:       "portwhine/certstream-trigger:latest",
			Version:     "1.0.0",
			OutputTypes: []string{"domain"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "domains": {"type": "array", "items": {"type": "string"}, "description": "Exact domains or wildcard patterns (*.example.com)"},
    "regex_patterns": {"type": "array", "items": {"type": "string"}, "description": "Go regex patterns to match against certificate domains"}
  }
}`,
			Icon:  "radio",
			Color: "#F59E0B",
		},
		{
			Id:          "ipaddress-trigger",
			DisplayName: "IP Address",
			Description: "Emits IP addresses from CIDRs, ranges, or single IPs on a configurable interval.",
			NodeType:    "trigger",
			Category:    "input",
			Image:       "portwhine/ipaddress-trigger:latest",
			Version:     "1.0.0",
			OutputTypes: []string{"ip_address"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "targets": {"type": "array", "items": {"type": "string"}, "description": "List of IPs, CIDRs, or IP ranges (start-end)"},
    "interval_seconds": {"type": "integer", "description": "Repetition interval in seconds (0 = emit once)"}
  },
  "required": ["targets"]
}`,
			Icon:  "globe",
			Color: "#10B981",
		},

		// ── Workers: Scanning ───────────────────────────────────────
		{
			Id:                 "nmap-worker",
			DisplayName:        "Nmap Scanner",
			Description:        "Port scanning and service fingerprinting via Nmap.",
			NodeType:           "worker",
			Category:           "scanning",
			Image:              "portwhine/nmap-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"ip_address", "domain"},
			OutputTypes:        []string{"service", "url"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "ports": {"type": "string", "description": "Port range (e.g. 1-1000, 80,443)", "default": "1-1000"},
    "scan_type": {"type": "string", "description": "Nmap scan type flag (e.g. -sT, -sS)", "default": "-sT"},
    "extra_args": {"type": "string", "description": "Additional nmap arguments"},
    "service_detection": {"type": "boolean", "description": "Enable -sV service fingerprinting", "default": true},
    "top_ports": {"type": "number", "description": "Scan top N ports (overrides ports)"},
    "timeout": {"type": "number", "description": "Per-host timeout in seconds", "default": 300},
    "max_retries": {"type": "number", "description": "Max retries per probe", "default": 1},
    "min_rate": {"type": "number", "description": "Min packets per second"},
    "timing": {"type": "string", "description": "Timing template (e.g. -T4)"}
  }
}`,
			Icon:  "radar",
			Color: "#3B82F6",
		},
		{
			Id:                 "ffuf-worker",
			DisplayName:        "FFUF Fuzzer",
			Description:        "Web content discovery and directory fuzzing.",
			NodeType:           "worker",
			Category:           "scanning",
			Image:              "portwhine/ffuf-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"url"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "wordlist": {"type": "string", "description": "Path to wordlist or 'common' for built-in", "default": "common"},
    "extensions": {"type": "string", "description": "File extensions to fuzz (e.g. '.php,.html')"},
    "threads": {"type": "number", "description": "Number of concurrent threads", "default": 40},
    "match_codes": {"type": "string", "description": "HTTP status codes to match", "default": "200,204,301,302,307,401,403"},
    "extra_args": {"type": "string", "description": "Additional ffuf command-line arguments"},
    "headers": {"type": "object", "description": "Custom HTTP headers"},
    "method": {"type": "string", "description": "HTTP method", "default": "GET"},
    "post_data": {"type": "string", "description": "POST body data"},
    "rate_limit": {"type": "number", "description": "Max requests/sec (0=unlimited)", "default": 0},
    "recursion": {"type": "boolean", "description": "Enable recursive fuzzing", "default": false},
    "recursion_depth": {"type": "number", "description": "Max recursion depth", "default": 2},
    "filter_size": {"type": "string", "description": "Filter by response size (-fs)"},
    "filter_words": {"type": "string", "description": "Filter by word count (-fw)"},
    "filter_lines": {"type": "string", "description": "Filter by line count (-fl)"},
    "filter_regex": {"type": "string", "description": "Filter by regex (-fr)"},
    "auto_calibrate": {"type": "boolean", "description": "Auto-calibrate filters (-ac)", "default": false},
    "proxy": {"type": "string", "description": "HTTP proxy URL"},
    "timeout": {"type": "number", "description": "Per-request timeout in seconds", "default": 10},
    "follow_redirects": {"type": "boolean", "description": "Follow HTTP redirects", "default": true}
  }
}`,
			Icon:  "search",
			Color: "#EF4444",
		},
		{
			Id:                 "nuclei-worker",
			DisplayName:        "Nuclei Scanner",
			Description:        "Template-based vulnerability scanning via Nuclei.",
			NodeType:           "worker",
			Category:           "scanning",
			Image:              "portwhine/nuclei-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url", "service"},
			OutputTypes:        []string{"vulnerability"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "templates": {"type": "string", "description": "Specific template paths/IDs to use (comma-separated)"},
    "severity": {"type": "string", "description": "Filter by severity (comma-separated)", "default": "medium,high,critical"},
    "tags": {"type": "string", "description": "Filter templates by tags (comma-separated)"},
    "rate_limit": {"type": "number", "description": "Max requests per second", "default": 150},
    "concurrency": {"type": "number", "description": "Number of concurrent templates", "default": 25},
    "timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 300},
    "extra_args": {"type": "string", "description": "Additional nuclei CLI arguments"}
  }
}`,
			Icon:  "shield-alert",
			Color: "#DC2626",
		},

		// ── Workers: Enumeration ────────────────────────────────────
		{
			Id:                 "subfinder-worker",
			DisplayName:        "Subfinder",
			Description:        "Subdomain enumeration using multiple sources.",
			NodeType:           "worker",
			Category:           "enumeration",
			Image:              "portwhine/subfinder-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"domain"},
			OutputTypes:        []string{"domain"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "sources": {"type": "string", "description": "Comma-separated list of sources to use"},
    "threads": {"type": "number", "description": "Number of concurrent threads", "default": 10},
    "timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 120},
    "extra_args": {"type": "string", "description": "Additional subfinder CLI arguments"},
    "recursive": {"type": "boolean", "description": "Enable recursive subdomain enumeration", "default": false},
    "max_depth": {"type": "number", "description": "Max recursion depth when recursive is true", "default": 3}
  }
}`,
			Icon:  "git-branch",
			Color: "#8B5CF6",
		},
		{
			Id:                 "resolver-worker",
			DisplayName:        "DNS Resolver",
			Description:        "Resolves domains to IPs and DNS records (A, AAAA, CNAME, MX, TXT, NS).",
			NodeType:           "worker",
			Category:           "enumeration",
			Image:              "portwhine/resolver-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"domain"},
			OutputTypes:        []string{"ip_address", "dns_record"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "dns_server": {"type": "string", "description": "Custom DNS server (e.g. 8.8.8.8:53)"},
    "record_types": {"type": "array", "items": {"type": "string"}, "description": "DNS record types to resolve (A, AAAA, CNAME, MX, TXT, NS)", "default": ["A", "AAAA"]},
    "timeout": {"type": "number", "description": "DNS lookup timeout in seconds", "default": 10},
    "retries": {"type": "number", "description": "Retry count on failure", "default": 2}
  }
}`,
			Icon:  "network",
			Color: "#6366F1",
		},
		{
			Id:                 "whois-worker",
			DisplayName:        "WHOIS Lookup",
			Description:        "Performs WHOIS lookups for domains and IP addresses.",
			NodeType:           "worker",
			Category:           "enumeration",
			Image:              "portwhine/whois-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"domain", "ip_address"},
			OutputTypes:        []string{"whois_result"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "server": {"type": "string", "description": "Specific WHOIS server to query"},
    "timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 30},
    "extra_args": {"type": "string", "description": "Additional whois CLI arguments"}
  }
}`,
			Icon:  "file-search",
			Color: "#14B8A6",
		},

		// ── Workers: Analysis ───────────────────────────────────────
		{
			Id:                 "screenshot-worker",
			DisplayName:        "Screenshot",
			Description:        "Captures browser screenshots of web pages.",
			NodeType:           "worker",
			Category:           "analysis",
			Image:              "portwhine/screenshot-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"screenshot"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "width": {"type": "number", "description": "Viewport width in pixels", "default": 1280},
    "height": {"type": "number", "description": "Viewport height in pixels", "default": 720},
    "timeout": {"type": "number", "description": "Navigation timeout in seconds", "default": 30},
    "full_page": {"type": "boolean", "description": "Capture full page screenshot", "default": false},
    "user_agent": {"type": "string", "description": "Custom User-Agent string"},
    "delay": {"type": "number", "description": "Seconds to wait after page load", "default": 2},
    "ignore_cert_errors": {"type": "boolean", "description": "Skip TLS certificate validation", "default": true},
    "device": {"type": "string", "description": "Device preset: 'mobile' (375x812) or 'tablet' (768x1024)"},
    "quality": {"type": "number", "description": "Screenshot quality (1-100)", "default": 90},
    "format": {"type": "string", "description": "Output format: 'png' or 'jpeg'", "default": "png"}
  }
}`,
			Icon:  "camera",
			Color: "#EC4899",
		},
		{
			Id:                 "testssl-worker",
			DisplayName:        "TestSSL",
			Description:        "Comprehensive SSL/TLS analysis using testssl.sh.",
			NodeType:           "worker",
			Category:           "analysis",
			Image:              "portwhine/testssl-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"service"},
			OutputTypes:        []string{"ssl_result"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "checks": {"type": "string", "description": "Specific testssl.sh checks to run (e.g. --protocols --ciphers)"},
    "extra_args": {"type": "string", "description": "Additional arguments to pass to testssl.sh"},
    "protocols_only": {"type": "boolean", "description": "Only test protocols (--protocols)", "default": false},
    "vulnerabilities_only": {"type": "boolean", "description": "Only test vulns (--vulnerable)", "default": false},
    "ciphers": {"type": "boolean", "description": "Test cipher suites (--ciphers)", "default": false},
    "headers": {"type": "boolean", "description": "Check HTTP headers (--headers)", "default": false},
    "sni": {"type": "string", "description": "Server Name Indication value"},
    "timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 300},
    "parallel": {"type": "boolean", "description": "Run checks in parallel (--parallel)", "default": false},
    "starttls": {"type": "string", "description": "STARTTLS protocol override (smtp, ftp, etc.)"}
  }
}`,
			Icon:  "lock",
			Color: "#F97316",
		},
		{
			Id:                 "ssh-audit-worker",
			DisplayName:        "SSH Audit",
			Description:        "Audits SSH server configuration and algorithms.",
			NodeType:           "worker",
			Category:           "analysis",
			Image:              "portwhine/ssh-audit-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"service"},
			OutputTypes:        []string{"ssh_audit_result"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 60},
    "extra_args": {"type": "string", "description": "Additional ssh-audit CLI arguments"},
    "policy": {"type": "string", "description": "Path to custom policy file for compliance checking"},
    "level": {"type": "string", "description": "Minimum output level: info, warn, fail", "default": "info"}
  }
}`,
			Icon:  "terminal",
			Color: "#A855F7",
		},
		{
			Id:                 "webanalyzer-worker",
			DisplayName:        "Web Analyzer",
			Description:        "Detects web technologies, frameworks, and CMS systems.",
			NodeType:           "worker",
			Category:           "analysis",
			Image:              "portwhine/webanalyzer-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"web_technology"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "timeout": {"type": "number", "description": "HTTP request timeout in seconds", "default": 30},
    "headers": {"type": "object", "description": "Custom HTTP headers"},
    "max_redirects": {"type": "number", "description": "Maximum redirects to follow", "default": 5},
    "user_agent": {"type": "string", "description": "Custom User-Agent string"},
    "ignore_cert_errors": {"type": "boolean", "description": "Skip TLS cert validation", "default": true}
  }
}`,
			Icon:  "layers",
			Color: "#06B6D4",
		},
		{
			Id:                 "humble-worker",
			DisplayName:        "HTTP Headers",
			Description:        "Analyzes HTTP security headers and best practices.",
			NodeType:           "worker",
			Category:           "analysis",
			Image:              "portwhine/humble-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"url"},
			OutputTypes:        []string{"http_headers"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "extra_args": {"type": "string", "description": "Additional command-line arguments for humble"},
    "skip_checks": {"type": "array", "items": {"type": "string"}, "description": "Checks to skip (--skip per entry)"},
    "headers": {"type": "object", "description": "Custom HTTP headers to send"},
    "timeout": {"type": "number", "description": "Execution timeout in seconds", "default": 120},
    "brief": {"type": "boolean", "description": "Brief output mode (-b)", "default": false},
    "user_agent": {"type": "string", "description": "Custom User-Agent string"}
  }
}`,
			Icon:  "file-code",
			Color: "#84CC16",
		},

		// ── Workers: Reporting ──────────────────────────────────────
		{
			Id:                 "report-worker",
			DisplayName:        "Report Generator",
			Description:        "Aggregates pipeline results into structured reports.",
			NodeType:           "worker",
			Category:           "reporting",
			Image:              "portwhine/report-worker:latest",
			Version:            "3.0.0",
			AcceptedInputTypes: []string{"service", "url", "vulnerability", "ssl_result", "http_headers", "web_technology", "screenshot", "ssh_audit_result", "whois_result", "ip_address", "dns_record", "domain"},
			OutputTypes:        []string{"report"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "format": {"type": "string", "description": "Output format: json or summary", "default": "json"},
    "include_raw": {"type": "boolean", "description": "Include raw payloads in report", "default": false},
    "group_by": {"type": "string", "description": "How to group results: type, host, or severity", "default": "type"},
    "title": {"type": "string", "description": "Report title", "default": "Portwhine Scan Report"}
  }
}`,
			Icon:  "file-text",
			Color: "#6B7280",
		},

		// ── Outputs (Integration/Notification) ──────────────────────
		{
			Id:                 "webhook-output",
			DisplayName:        "Webhook",
			Description:        "Sends pipeline results to a configured URL via HTTP POST.",
			NodeType:           "output",
			Category:           "output",
			Image:              "portwhine/webhook-output:latest",
			Version:            "1.0.0",
			AcceptedInputTypes: []string{"service", "url", "vulnerability", "ssl_result", "http_headers", "web_technology", "screenshot", "ssh_audit_result", "whois_result", "ip_address", "dns_record", "domain", "report"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "url": {"type": "string", "description": "Destination webhook URL"},
    "method": {"type": "string", "description": "HTTP method (POST or PUT)", "default": "POST"},
    "headers": {"type": "object", "description": "Custom HTTP headers (e.g. Authorization)"},
    "batch_size": {"type": "number", "description": "Number of items to batch per request", "default": 1},
    "timeout": {"type": "number", "description": "HTTP request timeout in seconds", "default": 30}
  },
  "required": ["url"]
}`,
			Icon:  "webhook",
			Color: "#336791",
		},
		{
			Id:                 "email-output",
			DisplayName:        "E-Mail Report",
			Description:        "Sends a summarized report of pipeline results via e-mail.",
			NodeType:           "output",
			Category:           "output",
			Image:              "portwhine/email-output:latest",
			Version:            "1.0.0",
			AcceptedInputTypes: []string{"service", "url", "vulnerability", "ssl_result", "http_headers", "web_technology", "screenshot", "ssh_audit_result", "whois_result", "ip_address", "dns_record", "domain", "report"},
			ConfigSchema: `{
  "type": "object",
  "properties": {
    "recipients": {"type": "array", "items": {"type": "string"}, "description": "E-mail addresses to send the report to"},
    "subject": {"type": "string", "description": "E-mail subject line", "default": "Portwhine Pipeline Report"},
    "format": {"type": "string", "description": "Report format: html or text", "default": "html"},
    "smtp_host": {"type": "string", "description": "SMTP server hostname"},
    "smtp_port": {"type": "number", "description": "SMTP server port", "default": 587},
    "smtp_user": {"type": "string", "description": "SMTP username"},
    "smtp_pass": {"type": "string", "description": "SMTP password"}
  },
  "required": ["recipients"]
}`,
			Icon:  "mail",
			Color: "#78716C",
		},
	}
}
