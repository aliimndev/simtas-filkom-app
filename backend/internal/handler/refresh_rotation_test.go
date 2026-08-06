//go:build integration

package handler_test

import (
	"net/http"
	"testing"

	"github.com/aliimndev/simtas-filkom-app/backend/internal/handler"
	"github.com/aliimndev/simtas-filkom-app/backend/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// TestRefreshTokenRotationLifecycle drives the refresh-token lifecycle through
// the real HTTP stack: login sets an HttpOnly cookie, refresh rotates it, and a
// replayed (stale) refresh token is rejected via reuse detection.
func TestRefreshTokenRotationLifecycle(t *testing.T) {
	c := lifecycleApp(t)
	_ = c.Do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    studentEmail,
		"password": studentPassword,
	}, "")

	// The HttpOnly refresh cookie must have been issued at login.
	first := c.CookieValue(handler.RefreshCookieName)
	assert.NotEmpty(t, first, "login must set the refresh-token cookie")

	// Refresh → rotated cookie (new family member), new access token.
	wRefresh := c.Do(http.MethodPost, "/api/v1/auth/refresh", nil, "")
	assert.Equal(t, http.StatusOK, wRefresh.Code, "refresh: %s", wRefresh.Body.String())
	second := c.CookieValue(handler.RefreshCookieName)
	assert.NotEmpty(t, second)
	assert.NotEqual(t, first, second, "refresh must rotate the cookie")

	// The rotated cookie must not appear in the JSON body (HttpOnly contract).
	var body struct {
		Data struct {
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	testutil.DecodeBody(t, wRefresh, &body)
	assert.Equal(t, "", body.Data.RefreshToken, "refresh token must stay in the cookie")

	// Refresh again with the current cookie → still valid.
	wRefresh2 := c.Do(http.MethodPost, "/api/v1/auth/refresh", nil, "")
	assert.Equal(t, http.StatusOK, wRefresh2.Code, "second refresh: %s", wRefresh2.Body.String())
	third := c.CookieValue(handler.RefreshCookieName)
	assert.NotEqual(t, second, third, "every refresh rotates the cookie")
}

// TestRefreshTokenReuseDetection verifies that replaying a rotated-out refresh
// token revokes the whole family: the stale token returns 401 and the current
// cookie is invalidated too (theft detection).
func TestRefreshTokenReuseDetection(t *testing.T) {
	c := lifecycleApp(t)
	_ = c.Do(http.MethodPost, "/api/v1/auth/login", map[string]string{
		"email":    studentEmail,
		"password": studentPassword,
	}, "")

	stale := c.CookieValue(handler.RefreshCookieName)

	// First refresh rotates the token; `stale` is now the previous member.
	wRefresh := c.Do(http.MethodPost, "/api/v1/auth/refresh", nil, "")
	assert.Equal(t, http.StatusOK, wRefresh.Code)
	current := c.CookieValue(handler.RefreshCookieName)
	assert.NotEqual(t, stale, current)

	// Replay the stale token: seed a fresh stateless client with the old cookie.
	replay := testutil.NewClient(c.Router())
	replay.SetCookie(handler.RefreshCookieName, stale)
	wReplay := replay.Do(http.MethodPost, "/api/v1/auth/refresh", nil, "")
	assert.Equal(t, http.StatusUnauthorized, wReplay.Code, "replayed refresh must be rejected: %s", wReplay.Body.String())

	// The whole family is revoked: even the current cookie can no longer refresh.
	wAfter := c.Do(http.MethodPost, "/api/v1/auth/refresh", nil, "")
	assert.Equal(t, http.StatusUnauthorized, wAfter.Code, "family must be revoked after reuse: %s", wAfter.Body.String())
}
