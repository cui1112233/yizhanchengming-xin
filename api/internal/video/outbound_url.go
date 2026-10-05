package video

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
)

var carrierGradeNAT = mustCIDR("100.64.0.0/10")

func mustCIDR(raw string) *net.IPNet {
	_, network, err := net.ParseCIDR(raw)
	if err != nil {
		panic(err)
	}
	return network
}

func unsafeOutboundIP(ip net.IP) bool {
	if ip == nil {
		return true
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsUnspecified() || ip.IsMulticast() || carrierGradeNAT.Contains(ip)
}

func loopbackHost(host string) bool {
	host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func validateRemoteMediaURL(raw string, allowInsecureLoopback bool) (*url.URL, error) {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" {
		return nil, fmt.Errorf("video: remote media URL is invalid")
	}
	host := parsed.Hostname()
	if host == "" {
		return nil, fmt.Errorf("video: remote media URL is invalid")
	}
	if allowInsecureLoopback && parsed.Scheme == "http" && loopbackHost(host) {
		return parsed, nil
	}
	if parsed.Scheme != "https" {
		return nil, fmt.Errorf("video: remote media URL must use https")
	}
	if loopbackHost(host) {
		return nil, fmt.Errorf("video: remote media URL targets a blocked address")
	}
	if ip := net.ParseIP(host); ip != nil && unsafeOutboundIP(ip) {
		return nil, fmt.Errorf("video: remote media URL targets a blocked address")
	}
	return parsed, nil
}

func safeDialContext(allowLoopback bool) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, fmt.Errorf("video: invalid outbound address")
		}
		addresses, err := net.DefaultResolver.LookupIPAddr(ctx, host)
		if err != nil || len(addresses) == 0 {
			return nil, fmt.Errorf("video: resolve outbound media host: %w", err)
		}
		var lastErr error
		for _, resolved := range addresses {
			ip := resolved.IP
			if unsafeOutboundIP(ip) {
				if !(allowLoopback && ip.IsLoopback()) {
					return nil, fmt.Errorf("video: outbound media host resolved to a blocked address")
				}
			}
			conn, dialErr := dialer.DialContext(ctx, network, net.JoinHostPort(ip.String(), port))
			if dialErr == nil {
				return conn, nil
			}
			lastErr = dialErr
		}
		if lastErr == nil {
			lastErr = fmt.Errorf("no safe outbound address")
		}
		return nil, lastErr
	}
}

func hardenedHTTPClient(base *http.Client, allowLoopback bool) *http.Client {
	if base == nil {
		base = &http.Client{}
	}
	clone := *base
	previousRedirect := clone.CheckRedirect
	clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if _, err := validateRemoteMediaURL(req.URL.String(), allowLoopback); err != nil {
			return err
		}
		if previousRedirect != nil {
			return previousRedirect(req, via)
		}
		if len(via) >= 10 {
			return fmt.Errorf("video: too many outbound redirects")
		}
		return nil
	}

	var transport *http.Transport
	switch current := clone.Transport.(type) {
	case nil:
		transport = http.DefaultTransport.(*http.Transport).Clone()
	case *http.Transport:
		transport = current.Clone()
	default:
		// Tests and explicitly injected transports retain their custom dial path;
		// URL literal and redirect validation still runs before they are invoked.
		return &clone
	}
	transport.DialContext = safeDialContext(allowLoopback)
	clone.Transport = transport
	return &clone
}
