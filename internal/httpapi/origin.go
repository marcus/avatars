package httpapi

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

// NormalizePublicURL returns the browser's canonical HTTPS origin for a proxy.
// It accepts no credentials, query, fragment, or path beyond the root.
func NormalizePublicURL(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" || u.Opaque != "" {
		return "", fmt.Errorf("public URL must be an HTTPS origin without a path, such as https://host.example:7447")
	}
	host := strings.ToLower(u.Hostname())
	if net.ParseIP(host) == nil {
		for _, c := range host {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-' || c == '.') {
				return "", fmt.Errorf("public URL requires an ASCII DNS hostname or IP address")
			}
		}
		if strings.Contains(host, "..") || strings.HasPrefix(host, ".") {
			return "", fmt.Errorf("public URL has an invalid hostname")
		}
	}
	port := u.Port()
	if port != "" {
		n, e := strconv.Atoi(port)
		if e != nil || n < 1 || n > 65535 {
			return "", fmt.Errorf("public URL port must be 1..65535")
		}
		port = strconv.Itoa(n)
	}
	if port == "443" {
		port = ""
	}
	if port != "" {
		host = net.JoinHostPort(host, port)
	} else if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	return "https://" + host, nil
}
