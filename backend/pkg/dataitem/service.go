package dataitem

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/types/known/structpb"
)

// Well-known service name constants matching nmap naming conventions.
const (
	ServiceHTTP      = "http"
	ServiceHTTPS     = "https"
	ServiceHTTPAlt   = "http-alt"
	ServiceHTTPSAlt  = "https-alt"
	ServiceHTTPProxy = "http-proxy"
	ServiceSSH       = "ssh"
	ServiceFTP       = "ftp"
	ServiceSMTP      = "smtp"
	ServiceSMTPS     = "smtps"
	ServiceIMAP      = "imap"
	ServiceIMAPS     = "imaps"
	ServicePOP3      = "pop3"
	ServicePOP3S     = "pop3s"
	ServiceMySQL     = "mysql"
	ServicePostgres  = "postgresql"
	ServiceRDP       = "ms-wbt-server"
	ServiceDNS       = "domain"
	ServiceSMB       = "microsoft-ds"
	ServiceTelnet    = "telnet"
	ServiceVNC       = "vnc"
	ServiceRedis     = "redis"
	ServiceMongoDB   = "mongodb"
)

// starttlsProtocols maps nmap service names to testssl.sh --starttls values.
var starttlsProtocols = map[string]string{
	ServiceSMTP: "smtp",
	ServiceFTP:  "ftp",
	ServiceIMAP: "imap",
	ServicePOP3: "pop3",
}

// IsHTTPServiceName returns true if the service name indicates an HTTP-family service.
func IsHTTPServiceName(serviceName string) bool {
	return IsHTTPService(serviceName)
}

// IsTLSService returns true if the service uses TLS/SSL wrapping.
// Checks for ssl/ prefix, https variants, and well-known TLS ports.
func IsTLSService(serviceName string, port int) bool {
	lower := strings.ToLower(serviceName)
	if strings.HasPrefix(lower, "ssl/") {
		return true
	}
	if strings.HasPrefix(lower, "https") {
		return true
	}
	if lower == ServiceSMTPS || lower == ServiceIMAPS || lower == ServicePOP3S {
		return true
	}
	if strings.Contains(lower, "ssl") || strings.Contains(lower, "tls") {
		return true
	}
	switch port {
	case 443, 8443, 993, 995, 465:
		return true
	}
	return false
}

// StarttlsProtocol returns the testssl.sh --starttls value for the given service name.
// Returns empty string if the service does not support STARTTLS.
func StarttlsProtocol(serviceName string) string {
	return starttlsProtocols[strings.ToLower(serviceName)]
}

// BuildServiceData constructs the data fields map for a service DataItem.
func BuildServiceData(ip string, port int, protocol, serviceName, serviceProduct, serviceVersion, hostname string) map[string]*structpb.Value {
	fields := map[string]*structpb.Value{
		"ip":              structpb.NewStringValue(ip),
		"port":            structpb.NewNumberValue(float64(port)),
		"protocol":        structpb.NewStringValue(protocol),
		"state":           structpb.NewStringValue("open"),
		"service_name":    structpb.NewStringValue(serviceName),
		"service_product": structpb.NewStringValue(serviceProduct),
		"service_version": structpb.NewStringValue(serviceVersion),
		"tls":             structpb.NewBoolValue(IsTLSService(serviceName, port)),
	}

	if hostname != "" {
		fields["hostname"] = structpb.NewStringValue(hostname)
	}

	if IsHTTPServiceName(serviceName) {
		scheme := DetermineScheme(serviceName, port)
		host := hostname
		if host == "" {
			host = ip
		}
		fields["scheme"] = structpb.NewStringValue(scheme)
		fields["url"] = structpb.NewStringValue(BuildURL(scheme, host, port))
	}

	return fields
}

// ExtractServiceURL extracts a URL from a service DataItem's data fields.
// Returns the URL and true if the service is HTTP-family, empty string and false otherwise.
func ExtractServiceURL(fields map[string]*structpb.Value) (string, bool) {
	if v, ok := fields["url"]; ok {
		if u := v.GetStringValue(); u != "" {
			return u, true
		}
	}

	// Fallback: derive URL from service fields if url field is missing.
	serviceName := ""
	if v, ok := fields["service_name"]; ok {
		serviceName = v.GetStringValue()
	}
	if !IsHTTPServiceName(serviceName) {
		return "", false
	}

	ip := ""
	if v, ok := fields["ip"]; ok {
		ip = v.GetStringValue()
	}
	if ip == "" {
		return "", false
	}

	port := 0
	if v, ok := fields["port"]; ok {
		port = int(v.GetNumberValue())
	}
	if port == 0 {
		return "", false
	}

	hostname := ip
	if v, ok := fields["hostname"]; ok {
		if h := v.GetStringValue(); h != "" {
			hostname = h
		}
	}

	scheme := DetermineScheme(serviceName, port)
	return BuildURL(scheme, hostname, port), true
}

// ExtractServiceHost returns the best hostname from a service DataItem's data fields.
// Prefers hostname over ip.
func ExtractServiceHost(fields map[string]*structpb.Value) string {
	if v, ok := fields["hostname"]; ok {
		if h := v.GetStringValue(); h != "" {
			return h
		}
	}
	if v, ok := fields["ip"]; ok {
		return v.GetStringValue()
	}
	return ""
}

// ExtractServicePort returns the port from a service DataItem's data fields.
func ExtractServicePort(fields map[string]*structpb.Value) int {
	if v, ok := fields["port"]; ok {
		return int(v.GetNumberValue())
	}
	return 0
}

// ExtractServiceName returns the service_name from a service DataItem's data fields.
func ExtractServiceName(fields map[string]*structpb.Value) string {
	if v, ok := fields["service_name"]; ok {
		return v.GetStringValue()
	}
	return ""
}

// FormatHostPort formats a host:port string.
func FormatHostPort(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}
