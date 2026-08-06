package httpapi

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"drone-platform/internal/domain"
)

// TokenManager issues and verifies bearer tokens.
//
// Two formats are supported:
//   - Legacy: base64(payload).base64(HMAC-SHA256) — current format
//   - JWT:    base64(header).base64(payload).base64(HMAC-SHA256) — standard JWT
//
// Issue() produces legacy tokens for backward compatibility.
// IssueJWT() produces standard JWT tokens (alg: HS256).
// Verify() accepts both formats transparently.
//
// Tokens expire after 15 minutes by default. Refresh tokens (stored separately)
// are used for long-lived sessions.
type TokenManager struct{ secret []byte }

// NewTokenManager creates a TokenManager. The secret must be at least 32 bytes.
func NewTokenManager(secret string) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, errors.New("AUTH_SECRET must be at least 32 bytes")
	}
	return &TokenManager{secret: []byte(secret)}, nil
}

// jwtHeader is the standard JWT header for HS256 tokens.
const jwtHeader = `{"alg":"HS256","typ":"JWT"}`

// sign produces a base64url-encoded HMAC-SHA256 signature over data.
func (m *TokenManager) sign(data string) string {
	h := hmac.New(sha256.New, m.secret)
	h.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(h.Sum(nil))
}

// Issue creates a signed bearer token in legacy format (no JWT header).
// Deprecated: use IssueJWT for new clients.
func (m *TokenManager) Issue(a domain.Actor, ttl time.Duration) (string, error) {
	p := struct {
		ID   string      `json:"sub"`
		Role domain.Role `json:"role"`
		Exp  int64       `json:"exp"`
	}{a.ID, a.Role, time.Now().Add(ttl).Unix()}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	body := base64.RawURLEncoding.EncodeToString(b)
	return body + "." + m.sign(body), nil
}

// IssueJWT creates a standard JWT token (HS256).
func (m *TokenManager) IssueJWT(a domain.Actor, ttl time.Duration) (string, error) {
	p := struct {
		ID   string      `json:"sub"`
		Role domain.Role `json:"role"`
		Exp  int64       `json:"exp"`
		Iat  int64       `json:"iat"`
	}{a.ID, a.Role, time.Now().Add(ttl).Unix(), time.Now().Unix()}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(jwtHeader))
	payloadB64 := base64.RawURLEncoding.EncodeToString(b)
	signingInput := headerB64 + "." + payloadB64
	return signingInput + "." + m.sign(signingInput), nil
}

// Verify validates a token's signature and expiry, returning the embedded Actor.
// Accepts both legacy tokens (2-part) and standard JWT tokens (3-part).
func (m *TokenManager) Verify(token string) (domain.Actor, error) {
	parts := strings.Split(token, ".")
	var payloadB64 string

	switch len(parts) {
	case 2:
		// Legacy format: payload.sig
		payloadB64 = parts[0]
		if !hmac.Equal([]byte(m.sign(parts[0])), []byte(parts[1])) {
			return domain.Actor{}, errors.New("invalid bearer token")
		}
	case 3:
		// Standard JWT: header.payload.sig
		payloadB64 = parts[1]
		signingInput := parts[0] + "." + parts[1]
		if !hmac.Equal([]byte(m.sign(signingInput)), []byte(parts[2])) {
			return domain.Actor{}, errors.New("invalid bearer token")
		}
	default:
		return domain.Actor{}, errors.New("invalid bearer token")
	}

	b, err := base64.RawURLEncoding.DecodeString(payloadB64)
	if err != nil {
		return domain.Actor{}, errors.New("invalid bearer token")
	}
	var p struct {
		ID   string      `json:"sub"`
		Role domain.Role `json:"role"`
		Exp  int64       `json:"exp"`
	}
	if err := json.Unmarshal(b, &p); err != nil || p.ID == "" || p.Exp <= time.Now().Unix() {
		return domain.Actor{}, errors.New("expired or invalid bearer token")
	}
	return domain.Actor{ID: p.ID, Role: p.Role}, nil
}

type actorKey struct{}

func (s *Server) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" || r.URL.Path == "/" ||
			r.URL.Path == "/admin" || r.URL.Path == "/favicon.ico" ||
			strings.HasPrefix(r.URL.Path, "/uploads/") ||
			strings.HasPrefix(r.URL.Path, "/swagger/") ||
			strings.HasPrefix(r.URL.Path, "/api/services/") ||
			r.URL.Path == "/api/v1/admin/token" ||
			strings.HasPrefix(r.URL.Path, "/api/v1/auth/") ||
			strings.HasPrefix(r.URL.Path, "/api/auth/") ||
			strings.HasPrefix(r.URL.Path, "/api/v1/webhooks/") {
			next.ServeHTTP(w, r)
			return
		}
		// Public read-only endpoints (GET, no auth required).
		// 公开前缀按路径匹配（如 /api/v1/jobs 会命中 /api/v1/jobs/mine），
		// 因此若请求携带有效 token 仍解析 actor 进 context，
		// 供 handler 区分登录态（jobs/mine、certificates/mine 等子路径依赖此行为）。
		if r.Method == http.MethodGet && isPublicPath(r.URL.Path) {
			if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
				if a, err := s.tokens.Verify(strings.TrimPrefix(h, "Bearer ")); err == nil {
					next.ServeHTTP(w, r.WithContext(contextWithActor(r, a)))
					return
				}
			}
			next.ServeHTTP(w, r)
			return
		}
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			fail(w, r, http.StatusUnauthorized, errors.New("bearer token required"))
			return
		}
		a, err := s.tokens.Verify(strings.TrimPrefix(h, "Bearer "))
		if err != nil {
			fail(w, r, http.StatusUnauthorized, err)
			return
		}
		next.ServeHTTP(w, r.WithContext(contextWithActor(r, a)))
	})
}
func contextWithActor(r *http.Request, a domain.Actor) context.Context {
	return context.WithValue(r.Context(), actorKey{}, a)
}
func authenticatedActor(r *http.Request) (domain.Actor, bool) {
	a, ok := r.Context().Value(actorKey{}).(domain.Actor)
	if !ok || a.ID == "" || a.Role == "" {
		return domain.Actor{}, false
	}
	return a, true
}

// adminGate enforces admin-only access on all /api/v1/admin/* routes.
// Must be wrapped inside authenticate so the actor is available.
// /api/v1/admin/token is the dev-mode token issuance endpoint and is exempt.
func (s *Server) adminGate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/v1/admin/") && r.URL.Path != "/api/v1/admin/token" {
			a, ok := authenticatedActor(r)
			if !ok || (a.Role != domain.RolePlatformAdmin && a.Role != domain.RoleAssociationAdmin) {
				fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// isPublicPath returns true for GET endpoints whose handler does NOT check auth.
// Only add paths where the handler itself allows anonymous access.
//
// Matching is exact for list endpoints and allows a single-level detail path
// (e.g. /api/v1/jobs/{id}). Nested sub-resources (e.g. /api/v1/demands/{id}/applications)
// are NOT public — the prefix bug previously exposed bidder info without auth.
func isPublicPath(path string) bool {
	// 名录公开，但「我的飞手状态」需要认证（前缀匹配会误放行 /mine）
	if path == "/api/v1/certified-pilots/mine" {
		return false
	}
	publicPrefixes := []string{
		"/api/v1/home",
		"/api/v1/search",
		"/api/v1/demands",
		"/api/v1/posts",
		"/api/v1/comments",
		"/api/v1/jobs",
		"/api/v1/listings",
		"/api/v1/training-courses",
		"/api/v1/instructors",
		"/api/v1/certified-pilots",
		"/api/v1/products",
		"/api/v1/articles",
		"/api/v1/venues",
		"/api/v1/reviews",
		"/api/v1/contract-templates",
		"/api/v1/image",
		"/api/v1/experts",
		"/api/v1/cases",
		"/api/v1/compliance-docs",
		"/api/v1/compliance-standards",
		"/api/v1/industry-reports",
		"/api/v1/portfolios",
		"/api/v1/achievements",
		"/api/v1/rd-challenges",
		"/api/v1/research-projects",
		"/api/v1/competitions",
		"/api/v1/events",
		"/api/v1/industry-resources",
		"/api/v1/emergency-resources",
		"/api/v1/training/courses",
		"/api/v1/training/certificates",
		"/api/v1/study/tours",
		"/api/v1/test/sites",
		"/api/v1/rd/challenges",
		"/api/v1/research/projects",
		"/api/v1/industry/reports",
		"/api/v1/industry/resources",
		"/api/v1/emergency/resources",
		"/api/v1/services",
		"/api/v1/resource-pools",
		"/api/v1/test-sites",
		"/api/v1/exhibitions",
		"/api/v1/transformations",
		"/api/v1/colleges",
		"/api/v1/cooperation-programs",
		"/api/v1/rescue-cases",
		"/api/v1/emergency-depts",
		"/api/v1/emergency-dispatches",
		"/api/v1/emergency-drills",
		"/api/v1/association-members",
		"/api/v1/enterprises/public",
	}
	for _, p := range publicPrefixes {
		if path == p {
			return true
		}
		if rest, ok := strings.CutPrefix(path, p+"/"); ok && !strings.Contains(rest, "/") {
			return true
		}
	}
	return false
}
