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

// ── IVDP server config ─────────────────────────────────────────────────────────
// Values come directly from the IVDP client's vite.config.js proxy targets.
//
// Auth server  — handles public (o.) activities (login, OTP, etc.)
// Base server  — handles restricted (r.) activities; requires Authorization: Bearer <token>
const (
	// Login (o. — public, no token needed)
	ivdpAuthURL      = "https://cs2.mkcl.org/35KZLAfko4jsE2bNyZlSJsYUbZ2/o/mql"
	ivdpActivityName = "GenerateLoginTokenUsingPassword"

	// SSO token validation (r. — restricted, token sent in Authorization header)
	// Any r. activity validates the bearer token; GetUsers is lightweight and
	// requires no mandatory request fields.
	// ↓ Change this activity name if the IVDP backend exposes a dedicated
	//   ValidateToken / CheckSession endpoint in future.
	ivdpBaseURL         = "https://cs2.mkcl.org/35KZlQvwtfak1Je7B2SUNgrN0ao/r/mql"
	ivdpSSOActivity     = "GetUsers"

	sessionTTL = 8 * time.Hour
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
// Body: { "username": "...", "password": "...", "captcha": "..." }
// Forwards to the IVDP MQL auth server and, on success, issues a local session token.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Captcha  string `json:"captcha"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	// Trim leading/trailing whitespace from all fields before any validation.
	username := strings.TrimSpace(req.Username)
	password := strings.TrimSpace(req.Password)
	captcha  := strings.TrimSpace(req.Captcha)

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}
	if captcha == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Please enter the captcha"})
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

// ── SSO Token Verification ────────────────────────────────────────────────────

// ivdpSSOUserResult is the shape of the user record that IVDP restricted
// activities may include inside their result payload.
// Fields are optional — not all activities return all of them.
type ivdpSSOUserResult struct {
	UserID   interface{} `json:"userId"`
	UserName string      `json:"userName"`
	Name     string      `json:"name"`
	Email    string      `json:"email"`
}

// SSOVerify handles GET /auth/sso-verify?token=<ivdp_token>
//
// Flow:
//  1. Accept the caller's IVDP bearer token as a query parameter.
//  2. Forward it to the IVDP Base server (restricted endpoint) as
//     Authorization: Bearer <token>. Any r. activity validates the token;
//     IVDP returns errorCode 1000 if the token is live, non-1000 otherwise.
//  3. On success → create an 8-hour local session and return our session token.
//  4. On failure → return 401 Unauthorized.
func (h *AuthHandler) SSOVerify(c *gin.Context) {
	ivdpToken := strings.TrimSpace(c.Query("token"))
	if ivdpToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "token query parameter is required",
		})
		return
	}

	// ── Build the IVDP restricted-endpoint request ─────────────────────────
	// The MQL payload wraps the activity name as a key; no data fields are
	// required for GetUsers — an empty object is enough.
	payload := map[string]interface{}{
		ivdpSSOActivity: map[string]interface{}{},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[SSO] failed to marshal IVDP payload: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSO service error"})
		return
	}

	httpReq, err := http.NewRequest(http.MethodPost, ivdpBaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("[SSO] failed to build request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSO service error"})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Service-Header", ivdpSSOActivity)
	httpReq.Header.Set("Authorization", "Bearer "+ivdpToken) // ← token validated here

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("[SSO] IVDP base server unreachable: %v", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "IVDP authentication service is currently unreachable — please try again",
		})
		return
	}
	defer resp.Body.Close()

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[SSO] failed to read IVDP response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "SSO service error"})
		return
	}

	// ── Parse IVDP MQL response ────────────────────────────────────────────
	// Expected shape: { "GetUsers": { "errorCode": 1000|0, "error": "...", "result": ... } }
	var mqlResp map[string]ivdpActivity
	if err := json.Unmarshal(rawBody, &mqlResp); err != nil {
		// If the IVDP server returned an HTTP 401 the body is often plain text.
		if resp.StatusCode == http.StatusUnauthorized {
			log.Printf("[SSO] IVDP returned 401 — token is invalid or expired")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "token is invalid or has expired"})
			return
		}
		log.Printf("[SSO] unexpected IVDP response (status %d): %s", resp.StatusCode, string(rawBody))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is invalid or has expired"})
		return
	}

	activity, ok := mqlResp[ivdpSSOActivity]
	if !ok {
		log.Printf("[SSO] IVDP response missing activity key %q: %s", ivdpSSOActivity, string(rawBody))
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is invalid or has expired"})
		return
	}

	// IVDP signals success with errorCode 1000 (or 0 in some versions).
	if activity.ErrorCode != 1000 && activity.ErrorCode != 0 {
		errMsg := activity.Error
		if errMsg == "" {
			errMsg = fmt.Sprintf("IVDP errorCode %d", activity.ErrorCode)
		}
		log.Printf("[SSO] IVDP rejected token: %s", errMsg)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "token is invalid or has expired"})
		return
	}

	// ── Token is valid — extract display name ─────────────────────────────
	// Try to pull a username from the activity result. GetUsers may return a
	// list or a single user object depending on the IVDP version.
	displayName := extractSSOUsername(activity.Result, ivdpToken)

	// ── Issue our own 8-hour session token ────────────────────────────────
	sessionToken, err := randomHex(32)
	if err != nil {
		log.Printf("[SSO] failed to generate session token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create session"})
		return
	}

	expiresAt := time.Now().Add(sessionTTL)
	sessionStore.Store(sessionToken, authSession{Username: displayName, ExpiresAt: expiresAt})
	log.Printf("[SSO] issued session for %q via IVDP SSO; expires %s",
		displayName, expiresAt.Format(time.RFC3339))

	c.JSON(http.StatusOK, gin.H{
		"token":      sessionToken,
		"username":   displayName,
		"expires_at": expiresAt.Format(time.RFC3339),
	})
}

// extractSSOUsername tries to pull a readable display name from the IVDP
// GetUsers result. The result may be nil, a single object, or a JSON array —
// we handle all three gracefully.
func extractSSOUsername(result *ivdpLoginResult, fallbackHint string) string {
	if result != nil {
		if result.Name != "" {
			return result.Name
		}
		if result.UserName != "" {
			return result.UserName
		}
	}
	// Last resort: return a sanitised prefix of the token so the session is
	// traceable in logs without exposing the full credential.
	if len(fallbackHint) >= 8 {
		return "sso:" + fallbackHint[:8] + "…"
	}
	return "SSO User"
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
