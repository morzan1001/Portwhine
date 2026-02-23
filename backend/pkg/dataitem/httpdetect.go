package dataitem

import (
	"fmt"
	"strings"
)

// httpServiceNames are nmap service names that indicate HTTP/HTTPS.
var httpServiceNames = map[string]bool{
	"http":       true,
	"http-alt":   true,
	"http-proxy": true,
	"https":      true,
	"https-alt":  true,
}

// IsHTTPService returns true if the nmap service name indicates an HTTP service.
func IsHTTPService(serviceName string) bool {
	lower := strings.ToLower(serviceName)
	if httpServiceNames[lower] {
		return true
	}
	if strings.HasPrefix(lower, "ssl/http") {
		return true
	}
	return false
}

// IsHTTPPort returns true if the port number is commonly associated with HTTP.
// Used as a fallback when service detection is unavailable.
func IsHTTPPort(port int) bool {
	switch port {
	case 80, 443, 8080, 8443, 8000, 8888, 3000, 5000, 9090:
		return true
	default:
		return false
	}
}

// DetermineScheme returns "https" or "http" based on the service name and port.
func DetermineScheme(serviceName string, port int) string {
	lower := strings.ToLower(serviceName)
	if strings.HasPrefix(lower, "ssl/") || strings.HasPrefix(lower, "https") {
		return "https"
	}
	if port == 443 || port == 8443 {
		return "https"
	}
	if strings.Contains(lower, "ssl") || strings.Contains(lower, "tls") {
		return "https"
	}
	return "http"
}

// BuildURL constructs a URL from scheme, hostname, and port.
// Default ports (80 for http, 443 for https) are omitted.
func BuildURL(scheme, hostname string, port int) string {
	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", scheme, hostname)
	}
	return fmt.Sprintf("%s://%s:%d", scheme, hostname, port)
}

