package middleware

import (
	"crypto/subtle"
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Headers the Vercel frontends (bidkan-client, bidkan-backoffice) send on every server-side call:
// all their requests leave from a few shared Vercel IPs, so without these every user would share
// one rate-limit bucket.
const (
	HeaderForwardedClientIP = "X-Bidkan-Client-IP"
	HeaderProxySecret       = "X-Bidkan-Proxy-Secret"
	clientIPLocalsKey       = "client_ip"
)

// ResolveClientIP works out the caller's real IP once per request (read it with ClientIP).
//   - Request from a private address (edge Caddy on the Docker network): use X-Real-IP, which
//     bidkan.caddy always overwrites with the connecting IP, so clients cannot spoof it.
//   - X-Bidkan-Proxy-Secret matches proxySecret: use X-Bidkan-Client-IP (browser IP seen by Vercel).
//
// An empty proxySecret turns the second rule off.
func ResolveClientIP(proxySecret string) fiber.Handler {
	secret := []byte(proxySecret)
	return func(c *fiber.Ctx) error {
		c.Locals(clientIPLocalsKey, resolveClientIP(c, secret))
		return c.Next()
	}
}

func resolveClientIP(c *fiber.Ctx, secret []byte) string {
	ip := c.IP()
	if peer := net.ParseIP(ip); peer != nil && (peer.IsPrivate() || peer.IsLoopback()) {
		if real := parseIP(c.Get("X-Real-IP")); real != "" {
			ip = real
		}
	}
	if len(secret) > 0 && subtle.ConstantTimeCompare([]byte(c.Get(HeaderProxySecret)), secret) == 1 {
		if fwd := parseIP(c.Get(HeaderForwardedClientIP)); fwd != "" {
			ip = fwd
		}
	}
	return ip
}

// ClientIP returns the IP found by ResolveClientIP, or the socket IP when that middleware did not run.
func ClientIP(c *fiber.Ctx) string {
	if ip, ok := c.Locals(clientIPLocalsKey).(string); ok && ip != "" {
		return ip
	}
	return c.IP()
}

func parseIP(raw string) string {
	ip := net.ParseIP(strings.TrimSpace(raw))
	if ip == nil {
		return ""
	}
	return ip.String()
}
