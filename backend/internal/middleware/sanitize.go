package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// htmlTagRegex matches HTML tags for stripping. This is a basic sanitizer —
// for production-grade HTML sanitization, consider using bluemonday or similar.
var htmlTagRegex = regexp.MustCompile(`<[^>]*>`)

// dangerousPatterns matches common XSS/injection patterns in plain text fields.
var dangerousPatterns = []string{
	"<script",
	"javascript:",
	"onerror=",
	"onload=",
	"onclick=",
	"onfocus=",
	"onmouseover=",
	"onsubmit=",
	"<iframe",
	"<object",
	"<embed",
	"<form",
	"<input",
	"<textarea",
	"<select",
}

// SanitizeMiddleware strips HTML tags and dangerous patterns from JSON request
// body fields that are commonly used for user-generated content (notes, titles,
// abstracts, comments, etc.). This prevents stored XSS without requiring the
// frontend to handle all sanitization.
//
// The middleware reads the request body, sanitizes all string values in the
// top-level JSON object, and rewrites the body for downstream handlers.
func SanitizeMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only sanitize POST, PUT, PATCH requests with JSON content type.
		method := c.Request.Method
		if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			c.Next()
			return
		}

		// Read the request body.
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		// Sanitize the body content.
		sanitized := sanitizeJSONBytes(body)
		if sanitized != nil {
			// Only rewrite if sanitization actually changed something.
			c.Request.Body = io.NopCloser(strings.NewReader(*sanitized))
		} else {
			// Restore the original body if no sanitization was needed.
			c.Request.Body = io.NopCloser(strings.NewReader(string(body)))
		}

		c.Next()
	}
}

// sanitizeJSONBytes performs a basic sanitization on JSON content by stripping
// HTML tags and dangerous patterns from string values. This is a pragmatic
// approach that handles the most common XSS vectors without requiring a full
// JSON parser.
func sanitizeJSONBytes(data []byte) *string {
	s := string(data)
	changed := false

	// Strip HTML tags.
	if htmlTagRegex.MatchString(s) {
		s = htmlTagRegex.ReplaceAllString(s, "")
		changed = true
	}

	// Strip dangerous patterns (case-insensitive check).
	lower := strings.ToLower(s)
	for _, pattern := range dangerousPatterns {
		if strings.Contains(lower, pattern) {
			// Remove the pattern (case-insensitive).
			s = removePattern(s, pattern)
			lower = strings.ToLower(s)
			changed = true
		}
	}

	if !changed {
		return nil
	}

	slog.Debug("sanitized request body",
		"method", "",
		"removed_html", true,
	)
	return &s
}

// removePattern removes all occurrences of a pattern from a string (case-insensitive).
func removePattern(s, pattern string) string {
	lower := strings.ToLower(s)
	for {
		idx := strings.Index(lower, pattern)
		if idx == -1 {
			break
		}
		end := idx + len(pattern)
		s = s[:idx] + s[end:]
		lower = strings.ToLower(s)
	}
	return s
}
