package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadRequiresAppEnv — C3: a missing APP_ENV must fail fast instead of
// silently defaulting to development mode (non-Secure cookies, exposed swagger,
// debug logging) in production.
func TestLoadRequiresAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "")
	assert.Panics(t, func() { Load() })
}

// TestValidateRejectsUnknownAppEnv — a typo in APP_ENV would otherwise bypass
// every production branch and silently run in development mode.
func TestValidateRejectsUnknownAppEnv(t *testing.T) {
	for _, env := range []string{"produkction", "prod", "production "} {
		c := &Config{AppEnv: env}
		assert.Panicsf(t, c.Validate, "APP_ENV=%q must be rejected", env)
	}
}
