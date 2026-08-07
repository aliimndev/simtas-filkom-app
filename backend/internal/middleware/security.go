package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds standard security headers to every response.
// Covers OWASP Secure Headers recommended set:
//   - X-Content-Type-Options: prevents MIME-sniffing
//   - X-Frame-Options: blocks clickjacking
//   - X-XSS-Protection: legacy XSS filter (kept for older browsers)
//   - Referrer-Policy: limits referrer leakage
//   - Content-Security-Policy: restricts resource loading to prevent XSS
//   - Strict-Transport-Security: forces HTTPS for 1 year
//   - Permissions-Policy: disables unnecessary browser features
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME-sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		// Prevent clickjacking
		c.Header("X-Frame-Options", "DENY")
		// Legacy XSS filter
		c.Header("X-XSS-Protection", "1; mode=block")
		// Limit referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		// Content Security Policy: allow only same-origin + self for scripts/styles;
		// allow images from data: and https: (for profile photos from external storage).
		c.Header("Content-Security-Policy",
			"default-src 'self'; "+
			"script-src 'self'; "+
			"style-src 'self' 'unsafe-inline'; "+
			"img-src 'self' data: https:; "+
			"font-src 'self' data:; "+
			"connect-src 'self'; "+
			"frame-ancestors 'none'; "+
			"base-uri 'self'; "+
			"form-action 'self'; "+
			"upgrade-insecure-requests")
		// Force HTTPS for 1 year (include subdomains)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		// Disable unnecessary browser features
		c.Header("Permissions-Policy",
			"camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		c.Next()
	}
}

// CORSMiddleware handles CORS with a comma-separated list of allowed origins
func CORSMiddleware(allowedOrigins string) gin.HandlerFunc {
	origins := strings.Split(allowedOrigins, ",")

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && isAllowedOrigin(origin, origins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-XSRF-TOKEN")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func isAllowedOrigin(origin string, allowed []string) bool {
	for _, o := range allowed {
		o = strings.TrimSpace(o)
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}
