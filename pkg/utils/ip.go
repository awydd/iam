package utils

import (
	"net"
	"net/http"
	"strings"
	"sync"
)

var (
	trustedMu      sync.RWMutex
	trustedProxies []*net.IPNet = make([]*net.IPNet, 0)
)

// 设置受信任的代理 CIDR 列表（配合 Nginx 使用）
func SetTrustedProxies(cidrs []string) {
	nets := make([]*net.IPNet, 0, len(cidrs))
	for _, c := range cidrs {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}

		if !strings.Contains(c, "/") {
			if ip := net.ParseIP(c); ip != nil {
				if ip.To4() != nil {
					c += "/32"
				} else {
					c += "/128"
				}
			}
		}

		_, ipNet, err := net.ParseCIDR(c)
		if err == nil {
			nets = append(nets, ipNet)
		}
	}

	trustedMu.Lock()
	trustedProxies = nets
	trustedMu.Unlock()
}

func isTrustedProxy(ip net.IP) bool {
	if ip == nil {
		return false
	}
	trustedMu.RLock()
	defer trustedMu.RUnlock()

	for _, n := range trustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}

// 获取真实的客户端 IP
func ClientIP(r *http.Request) string {
	remoteIP := remoteAddrIP(r.RemoteAddr)

	if remoteIP == nil || !isTrustedProxy(remoteIP) {
		if remoteIP != nil {
			return remoteIP.String()
		}
		return r.RemoteAddr
	}

	if ip := clientIPFromXFF(r.Header.Get("X-Forwarded-For")); ip != "" {
		return ip
	}

	if xrip := strings.TrimSpace(r.Header.Get("X-Real-IP")); xrip != "" {
		if ip := net.ParseIP(xrip); ip != nil {
			return ip.String()
		}
	}

	return remoteIP.String()
}

func clientIPFromXFF(xff string) string {
	if xff == "" {
		return ""
	}
	parts := strings.Split(xff, ",")

	for i := len(parts) - 1; i >= 0; i-- {
		candidate := strings.TrimSpace(parts[i])
		ip := net.ParseIP(candidate)
		if ip == nil {
			continue
		}

		if !isTrustedProxy(ip) {
			return ip.String()
		}
	}
	return ""
}

func remoteAddrIP(remoteAddr string) net.IP {
	host, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		host = remoteAddr
	}
	return net.ParseIP(host)
}

func UserAgent(r *http.Request) string {
	return strings.TrimSpace(r.UserAgent())
}
