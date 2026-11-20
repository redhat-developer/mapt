package gitlab

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
)

// hostPattern matches a host name or a domain suffix such as ".example.com". Labels may
// contain underscores (_svc.internal), and a fully qualified name may end in a dot.
var hostPattern = regexp.MustCompile(`^\.?[A-Za-z0-9_]([A-Za-z0-9_-]*[A-Za-z0-9_])?(\.[A-Za-z0-9_]([A-Za-z0-9_-]*[A-Za-z0-9_])?)*\.?$`)

const encodeHint = "percent-encode characters such as @ : / ? # in a user name or password"

// ProxyEnv returns the proxy variables for the runner's job environment. Programs read
// different names, so each variable is set in upper and lower case. It returns nil when
// httpProxy is empty.
func ProxyEnv(httpProxy, noProxy string) ([]string, error) {
	if httpProxy == "" {
		if noProxy != "" {
			return nil, fmt.Errorf("--glrunner-no-proxy needs --glrunner-http-proxy")
		}
		return nil, nil
	}
	if err := validateProxyURL(httpProxy); err != nil {
		return nil, err
	}
	env := []string{
		"HTTP_PROXY=" + httpProxy,
		"http_proxy=" + httpProxy,
		"HTTPS_PROXY=" + httpProxy,
		"https_proxy=" + httpProxy,
	}
	if noProxy != "" {
		entries, err := noProxyEntries(noProxy)
		if err != nil {
			return nil, err
		}
		list := strings.Join(entries, ",")
		env = append(env, "NO_PROXY="+list, "no_proxy="+list)
	}
	return env, nil
}

// validateProxyURL checks the proxy URL. Errors show the URL without its password,
// or not at all when it does not parse.
func validateProxyURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return fmt.Errorf("--glrunner-http-proxy must be a URL such as http://proxy.example.com:3128; %s", encodeHint)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("--glrunner-http-proxy must use http or https, got %q", u.Redacted())
	}
	if (u.Path != "" && u.Path != "/") || u.RawQuery != "" || u.Fragment != "" {
		return fmt.Errorf("--glrunner-http-proxy must not have a path, query or fragment, got %q; %s", u.Redacted(), encodeHint)
	}
	return nil
}

// noProxyEntries splits and checks a NO_PROXY list. It accepts host names, domains, IP
// literals and "*". A CIDR is rejected because some clients, such as curl 7.76.1 on
// RHEL 9, ignore it without a message and send the request to the proxy. Repeated
// entries are dropped, keeping the first.
func noProxyEntries(list string) ([]string, error) {
	var entries []string
	seen := map[string]bool{}
	for _, raw := range strings.Split(list, ",") {
		entry := strings.TrimSpace(raw)
		if entry == "" || seen[strings.ToLower(entry)] {
			continue
		}
		seen[strings.ToLower(entry)] = true
		if _, _, err := net.ParseCIDR(entry); err == nil {
			return nil, fmt.Errorf("--glrunner-no-proxy entry %q is a CIDR range, which some clients (such as curl on RHEL 9) ignore without an error; list the host names or IP addresses instead", entry)
		}
		if entry != "*" && net.ParseIP(strings.Trim(entry, "[]")) == nil && !hostPattern.MatchString(entry) {
			return nil, fmt.Errorf("--glrunner-no-proxy entry %q is not a host name, domain, IP address or *", entry)
		}
		entries = append(entries, entry)
	}
	if len(entries) == 0 {
		return nil, fmt.Errorf("--glrunner-no-proxy has no entries, got %q", list)
	}
	return entries, nil
}

// shellQuote quotes a value for bash.
func shellQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", `'\''`) + "'"
}

// powershellQuote quotes a value for PowerShell.
func powershellQuote(v string) string {
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}
