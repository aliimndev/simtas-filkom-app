package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	// csrfCookieName is the name of the CSRF token cookie.
	csrfCookieName = "XSRF-TOKEN"
	// csrfHeaderName is the header name the frontend must send.
	csrfHeaderName = "X-XSRF-TOKEN"
	// csrfTokenLength is the length of the random CSRF token in bytes.
	csrfTokenLength = 32
)

// CSRFMiddleware implements the Double Submit Cookie pattern for CSRF protection.
//
// How it works:
//  1. On any GET request, the middleware generates a cryptographically random
//     CSRF token and sets it as a cookie (SameSite=Lax, HttpOnly=false so the
//     frontend JavaScript can read it).
//  2. The frontend reads the cookie value and includes it in the X-XSRF-TOKEN
//     header for every state-changing request (POST/PUT/PATCH/DELETE).
//  3. The middleware compares the cookie value with the header value. If they
//     don't match, the request is rejected with 403.
//
// Safe methods (GET, HEAD, OPTIONS) are exempt from CSRF checks because they
// are idempotent and cannot cause state changes.
//
// exemptPaths lists state-changing routes that skip the check. These are the
// public, unauthenticated auth endpoints (login, refresh, forgot/reset
// password): a fresh browser — or an emailed reset-password deep link — POSTs
// to them before any GET has seeded the CSRF cookie, so the Double Submit
// pattern cannot be established in advance. They remain CSRF-safe: the
// SameSite=Lax cookies are never sent on cross-site requests, and reset
// password is additionally guarded by the emailed reset token in the body.
//
// This approach works without server-side token storage and is compatible with
// SPA architectures.
func CSRFMiddleware(exemptPaths ...string) gin.HandlerFunc {
	// ponytail: exact registered route match (c.FullPath), so an unmatched path
	// can never be spoofed into the exempt list. Add new unauthenticated POST
	// routes here; keep the list as small as possible.
	exempt := make(map[string]struct{}, len(exemptPaths))
	for _, p := range exemptPaths {
		exempt[p] = struct{}{}
	}
	return func(c *gin.Context) {
		// Safe methods are exempt from CSRF checks.
		method := c.Request.Method
		if method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions {
			// Set CSRF cookie on safe methods if not present.
			if _, err := c.Cookie(csrfCookieName); err != nil {
				token := generateCSRFToken()
				c.SetSameSite(http.SameSiteLaxMode)
				c.SetCookie(
					csrfCookieName,
					token,
					int(24*time.Hour.Seconds()), // 24 hour expiry
					"/",
					"",   // domain (same as request)
					false, // secure — set to true behind HTTPS reverse proxy
					false, // HttpOnly: false so frontend JS can read it
				)
			}
			c.Next()
			return
		}

		// State-changing methods on an exempt route skip the check.
		if _, ok := exempt[c.FullPath()]; ok {
			c.Next()
			return
		}

		// State-changing methods: verify CSRF token.
		cookieVal, err := c.Cookie(csrfCookieName)
		if err != nil || cookieVal == "" {
			// No CSRF cookie — reject.
			abortCSRF(c, "CSRF token cookie tidak ditemukan")
			return
		}

		headerVal := c.GetHeader(csrfHeaderName)
		if headerVal == "" {
			// No CSRF header — reject.
			abortCSRF(c, "CSRF header tidak ditemukan")
			return
		}

		// Constant-time comparison to prevent timing attacks.
		if !secureCompare(cookieVal, headerVal) {
			abortCSRF(c, "CSRF token tidak cocok")
			return
		}

		c.Next()
	}
}

// abortCSRF sends a 403 response and aborts the request chain.
func abortCSRF(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
		"success": false,
		"message": message,
	})
}

// generateCSRFToken creates a cryptographically random CSRF token.
func generateCSRFToken() string {
	b := make([]byte, csrfTokenLength)
	if _, err := rand.Read(b); err != nil {
		// Fallback — should never happen with crypto/rand.
		return hex.EncodeToString([]byte("csrf-fallback-token"))
	}
	return hex.EncodeToString(b)
}

// secureCompare performs a constant-time string comparison to prevent timing attacks.
// Uses crypto/subtle for guaranteed constant-time comparison.
func secureCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

// GetCSRFHeaderName returns the header name for the CSRF token (for frontend use).
func GetCSRFHeaderName() string {
	return csrfHeaderName
}

// GetCSRFCookieName returns the cookie name for the CSRF token (for frontend use).
func GetCSRFCookieName() string {
	return csrfCookieName
}

// parseCSRFToken extracts the CSRF token from the cookie string.
func parseCSRFToken(cookieHeader string) string {
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, csrfCookieName+"=") {
			return strings.TrimPrefix(part, csrfCookieName+"=")
		}
	}
	return ""
}
