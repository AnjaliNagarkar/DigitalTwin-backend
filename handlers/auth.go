package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ── IVDP Auth-server config ────────────────────────────────────────────────────
// These values come directly from the IVDP client's vite.config.js proxy target.
// The auth server speaks the MKCL MQL protocol; we only call the public (o.) endpoint.
const (
	ivdpAuthURL      = "https://cs2.mkcl.org/35KZLAfko4jsE2bNyZlSJsYUbZ2/o/mql"
	ivdpActivityName = "GenerateLoginTokenUsingPassword"
	sessionTTL       = 8 * time.Hour
)

// ── Session store ──────────────────────────────────────────────────────────────

type authSession struct {
	Username  string
	ExpiresAt time.Time
}

var sessionStore sync.Map // token (string) → authSession

// ── IVDP MQL response shapes ──────────────────────────────────────────────────

// ivdpLoginResult mirrors the `result` object returned by
// GenerateLoginTokenUsingPassword when errorCode == 1000.
type ivdpLoginResult struct {
	Token          string      `json:"token"`
	UserID         interface{} `json:"userId"`
	UserName       string      `json:"userName"`
	Name           string      `json:"name"`
	EntityUserID   interface{} `json:"entityUserId"`
	DistrictID     interface{} `json:"districtId"`
	TalukaID       interface{} `json:"talukaId"`
	VillageID      interface{} `json:"villageId"`
	GrampanchayatID interface{} `json:"grampanchayatId"`
	Roles          string      `json:"roles"`
	ClientIp       string      `json:"clientIp"`
}

type ivdpActivity struct {
	ErrorCode int              `json:"errorCode"`
	Error     string           `json:"error"`
	Result    *ivdpLoginResult `json:"result"`
}

// ── AuthHandler ───────────────────────────────────────────────────────────────

// AuthHandler delegates credential verification to the IVDP auth server.
// It does not touch the local database.
type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	log.Printf("[AUTH] IVDP auth server: %s", ivdpAuthURL)
	return &AuthHandler{}
}

// Login handles POST /auth/login.
// Body: { "username": "...", "password": "..." }
// Forwards to the IVDP MQL auth server and, on success, issues a local session token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	if username == "" || password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password cannot be blank"})
		return
	}

	// ── Call IVDP auth server ──────────────────────────────────────────────────
	// MQL payload format: { "ActivityName": { "field": "value", ... } }
	payload := map[string]interface{}{
		ivdpActivityName: map[string]string{
			"userName": username,
			"password": password,
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[AUTH] failed to marshal IVDP payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication service error"})
		return
	}

	httpReq, err := http.NewRequest(http.MethodPost, ivdpAuthURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[AUTH] failed to create IVDP request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication service unavailable"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Service-Header", ivdpActivityName)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[AUTH] IVDP auth server unreachable: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "authentication service is currently unreachable — please try again"})
		return
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[AUTH] failed to read IVDP response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication service error"})
		return
	}

	// ── Parse IVDP MQL response ────────────────────────────────────────────────
	// Shape: { "GenerateLoginTokenUsingPassword": { errorCode, error, result: {...} } }
	var mqlResp map[string]ivdpActivity
	if err := json.Unmarshal(rawBody, &mqlResp); err != nil {
		log.Printf("[AUTH] unexpected IVDP response format: %s", string(rawBody))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication service returned an unexpected response"})
		return
	}

	activity, ok := mqlResp[ivdpActivityName]
	if !ok {
		log.Printf("[AUTH] IVDP response missing activity key %q: %s", ivdpActivityName, string(rawBody))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "authentication service error"})
		return
	}

	// IVDP uses errorCode 1000 (or 0) to signal success.
	if activity.ErrorCode != 1000 && activity.ErrorCode != 0 {
		errMsg := activity.Error
		if errMsg == "" {
			errMsg = fmt.Sprintf("login failed (code %d)", activity.ErrorCode)
		}
		log.Printf("[AUTH] IVDP rejected login for %q: %s", username, errMsg)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if activity.Result == nil {
		log.Printf("[AUTH] IVDP returned success but empty result for %q", username)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	// ── Issue our own session token ────────────────────────────────────────────
	sessionToken, err := randomHex(32)
	if err != nil {
		log.Printf("[AUTH] failed to generate session token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}

	expiresAt := time.Now().Add(sessionTTL)
	displayName := activity.Result.Name
	if displayName == "" {
		displayName = activity.Result.UserName
	}
	if displayName == "" {
		displayName = username
	}

	sessionStore.Store(sessionToken, authSession{Username: displayName, ExpiresAt: expiresAt})
	log.Printf("[AUTH] user %q (%s) logged in via IVDP; session expires %s",
		username, displayName, expiresAt.Format(time.RFC3339))

	c.JSON(http.StatusOK, gin.H{
		"token":      sessionToken,
		"username":   displayName,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// Logout handles POST /auth/logout — invalidates the caller's session token.
func (h *AuthHandler) Logout(c *gin.Context) {
	token := extractBearerToken(c)
	if token != "" {
		if val, ok := sessionStore.Load(token); ok {
			sess := val.(authSession)
			log.Printf("[AUTH] user %q logged out", sess.Username)
		}
		sessionStore.Delete(token)
	}
	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

// Me handles GET /auth/me — returns the currently logged-in username.
func (h *AuthHandler) Me(c *gin.Context) {
	username, _ := c.Get("username")
	c.JSON(http.StatusOK, gin.H{"username": username})
}

// ── Auth middleware ────────────────────────────────────────────────────────────

// AuthMiddleware rejects requests that do not carry a valid session token.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearerToken(c)
		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required — please log in"})
			return
		}

		val, ok := sessionStore.Load(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session not found — please log in again"})
			return
		}

		sess := val.(authSession)
		if time.Now().After(sess.ExpiresAt) {
			sessionStore.Delete(token)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired — please log in again"})
			return
		}

		// Sliding expiry: refresh TTL on activity.
		sess.ExpiresAt = time.Now().Add(sessionTTL)
		sessionStore.Store(token, sess)

		c.Set("username", sess.Username)
		c.Next()
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func extractBearerToken(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer ")
	}
	return ""
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
