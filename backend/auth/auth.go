package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"backend/database"
)

const (
	sessionCookieName = "repair_session"
	sessionDuration   = 8 * time.Hour
)

type sessionClaims struct {
	UserID    int   `json:"uid"`
	ExpiresAt int64 `json:"exp"`
}

var (
	signingKeyMu sync.RWMutex
	signingKey   []byte
)

func InitializeSessionKey() error {
	key := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if key == "" {
		if strings.EqualFold(os.Getenv("GIN_MODE"), gin.ReleaseMode) {
			return errors.New("SESSION_SECRET must be set when GIN_MODE=release")
		}

		randomKey := make([]byte, 32)
		if _, err := rand.Read(randomKey); err != nil {
			return fmt.Errorf("generate development session secret: %w", err)
		}
		key = string(randomKey)
		log.Print("SESSION_SECRET is unset; using an ephemeral development key (sessions expire when the backend restarts)")
	} else if len([]byte(key)) < 32 {
		return errors.New("SESSION_SECRET must contain at least 32 bytes")
	}

	signingKeyMu.Lock()
	signingKey = []byte(key)
	signingKeyMu.Unlock()
	return nil
}

func CreateSessionCookie(c *gin.Context, userID int) error {
	if userID <= 0 {
		return errors.New("invalid user ID")
	}

	expiresAt := time.Now().Add(sessionDuration)
	claims, err := json.Marshal(sessionClaims{UserID: userID, ExpiresAt: expiresAt.Unix()})
	if err != nil {
		return fmt.Errorf("encode session claims: %w", err)
	}

	payload := base64.RawURLEncoding.EncodeToString(claims)
	signature, err := sign(payload)
	if err != nil {
		return err
	}

	c.SetSameSite(sessionSameSite())
	c.SetCookie(
		sessionCookieName,
		payload+"."+base64.RawURLEncoding.EncodeToString(signature),
		int(sessionDuration.Seconds()),
		"/",
		"",
		c.Request.TLS != nil || strings.EqualFold(os.Getenv("SESSION_COOKIE_SECURE"), "true") || strings.EqualFold(os.Getenv("GIN_MODE"), gin.ReleaseMode),
		true,
	)
	return nil
}

func ClearSessionCookie(c *gin.Context) {
	c.SetSameSite(sessionSameSite())
	c.SetCookie(
		sessionCookieName,
		"",
		-1,
		"/",
		"",
		c.Request.TLS != nil || strings.EqualFold(os.Getenv("SESSION_COOKIE_SECURE"), "true") || strings.EqualFold(os.Getenv("GIN_MODE"), gin.ReleaseMode),
		true,
	)
}

func sessionSameSite() http.SameSite {
	if strings.EqualFold(os.Getenv("GIN_MODE"), gin.ReleaseMode) {
		return http.SameSiteNoneMode
	}
	switch strings.ToLower(strings.TrimSpace(os.Getenv("SESSION_COOKIE_SAMESITE"))) {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}

func UserID(c *gin.Context) int {
	value, exists := c.Get("auth_user_id")
	if !exists {
		return 0
	}
	userID, ok := value.(int)
	if !ok {
		return 0
	}
	return userID
}

func Role(c *gin.Context) string {
	value, _ := c.Get("auth_user_role")
	role, _ := value.(string)
	return role
}

func RequireRoles(allowedRoles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		allowed[role] = struct{}{}
	}

	return func(c *gin.Context) {
		cookie, err := c.Cookie(sessionCookieName)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "กรุณาเข้าสู่ระบบ"})
			return
		}

		userID, err := verify(cookie)
		if err != nil {
			ClearSessionCookie(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "เซสชันหมดอายุหรือไม่ถูกต้อง กรุณาเข้าสู่ระบบใหม่"})
			return
		}

		var role string
		err = database.DB.QueryRow(`
			SELECT role FROM users WHERE id = $1 AND is_active = TRUE
		`, userID).Scan(&role)
		if err == sql.ErrNoRows {
			ClearSessionCookie(c)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "บัญชีนี้ไม่สามารถใช้งานได้"})
			return
		}
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ไม่สามารถตรวจสอบสิทธิ์ผู้ใช้ได้"})
			return
		}
		if len(allowed) > 0 {
			if _, ok := allowed[role]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "คุณไม่มีสิทธิ์ใช้งานรายการนี้"})
				return
			}
		}

		c.Set("auth_user_id", userID)
		c.Set("auth_user_role", role)
		c.Next()
	}
}

func ProtectMutations(allowedOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		if origin = strings.TrimRight(strings.TrimSpace(origin), "/"); origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}

		origin := strings.TrimRight(strings.TrimSpace(c.GetHeader("Origin")), "/")
		if origin != "" {
			if _, ok := allowed[origin]; !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "ไม่อนุญาตให้เรียก API จากเว็บไซต์นี้"})
				return
			}
		} else if _, err := c.Cookie(sessionCookieName); err == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "คำขอที่ใช้เซสชันต้องระบุ Origin"})
			return
		}

		c.Next()
	}
}

func sign(payload string) ([]byte, error) {
	signingKeyMu.RLock()
	key := append([]byte(nil), signingKey...)
	signingKeyMu.RUnlock()
	if len(key) < 32 {
		return nil, errors.New("session signing key has not been initialized")
	}

	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(payload)); err != nil {
		return nil, fmt.Errorf("sign session: %w", err)
	}
	return mac.Sum(nil), nil
}

func verify(token string) (int, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, errors.New("invalid session format")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, errors.New("invalid session signature")
	}
	expected, err := sign(parts[0])
	if err != nil || !hmac.Equal(signature, expected) {
		return 0, errors.New("invalid session signature")
	}

	claimsJSON, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, errors.New("invalid session claims")
	}
	var claims sessionClaims
	if err := json.Unmarshal(claimsJSON, &claims); err != nil || claims.UserID <= 0 {
		return 0, errors.New("invalid session claims")
	}
	if time.Now().Unix() >= claims.ExpiresAt {
		return 0, errors.New("session expired")
	}
	return claims.UserID, nil
}
