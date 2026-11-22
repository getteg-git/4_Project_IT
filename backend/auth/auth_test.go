package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setTestSigningKey(t *testing.T) {
	t.Helper()
	signingKeyMu.Lock()
	original := signingKey
	signingKey = []byte("test-session-signing-key-that-is-at-least-32-bytes")
	signingKeyMu.Unlock()
	t.Cleanup(func() {
		signingKeyMu.Lock()
		signingKey = original
		signingKeyMu.Unlock()
	})
}

func TestVerifySession(t *testing.T) {
	setTestSigningKey(t)

	claims, err := json.Marshal(sessionClaims{UserID: 42, ExpiresAt: time.Now().Add(time.Minute).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString(claims)
	signature, err := sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	userID, err := verify(payload + "." + base64.RawURLEncoding.EncodeToString(signature))
	if err != nil || userID != 42 {
		t.Fatalf("verify() = (%d, %v), want (42, nil)", userID, err)
	}
	if _, err := verify(payload + ".invalid"); err == nil {
		t.Fatal("verify() accepted a modified signature")
	}
}

func TestVerifyRejectsExpiredSession(t *testing.T) {
	setTestSigningKey(t)

	claims, err := json.Marshal(sessionClaims{UserID: 42, ExpiresAt: time.Now().Add(-time.Minute).Unix()})
	if err != nil {
		t.Fatal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString(claims)
	signature, err := sign(payload)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := verify(payload + "." + base64.RawURLEncoding.EncodeToString(signature)); err == nil {
		t.Fatal("verify() accepted an expired session")
	}
}

func TestCreateSessionCookieUsesHttpOnlyAndStrictSameSite(t *testing.T) {
	setTestSigningKey(t)
	t.Setenv("GIN_MODE", "debug")
	t.Setenv("SESSION_COOKIE_SECURE", "false")
	t.Setenv("SESSION_COOKIE_SAMESITE", "strict")

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	if err := CreateSessionCookie(context, 42); err != nil {
		t.Fatal(err)
	}

	response := recorder.Result()
	defer response.Body.Close()
	cookies := response.Cookies()
	if len(cookies) != 1 {
		t.Fatalf("got %d cookies, want 1", len(cookies))
	}
	cookie := cookies[0]
	if cookie.Name != sessionCookieName || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("unexpected session cookie settings: %+v", cookie)
	}
}

func TestCreateSessionCookieUsesSecureCrossSiteSettingsInRelease(t *testing.T) {
	setTestSigningKey(t)
	t.Setenv("GIN_MODE", gin.ReleaseMode)
	t.Setenv("SESSION_COOKIE_SECURE", "false")

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodPost, "/api/login", nil)
	if err := CreateSessionCookie(context, 42); err != nil {
		t.Fatal(err)
	}

	cookie := recorder.Result().Cookies()[0]
	if cookie.SameSite != http.SameSiteNoneMode || !cookie.Secure || !cookie.HttpOnly {
		t.Fatalf("unexpected release session cookie settings: %+v", cookie)
	}
}

func TestProtectMutationsRejectsUntrustedOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ProtectMutations([]string{"http://localhost:5173"}))
	router.POST("/api/change", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	request := httptest.NewRequest(http.MethodPost, "/api/change", nil)
	request.Header.Set("Origin", "https://untrusted.example")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}
