package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MaxBodySize returns a middleware that caps the request body to maxBytes using
// http.MaxBytesReader. Once the limit is exceeded Read returns an error and the
// handler receives a 413 (Payload Too Large) — preventing unbounded payloads from
// exhausting server memory.
func MaxBodySize(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
