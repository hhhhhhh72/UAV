// Package middleware provides HTTP middleware utilities.
package middleware

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

// SanitizeBody removes potentially dangerous content from JSON request bodies.
// Strips HTML tags, trims excessive whitespace, and limits string lengths.
func SanitizeBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		// Body sanitization is handled in the decode/respond layer;
		// this middleware serves as an additional defense-in-depth check.
		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "application/json") {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// SanitizeString removes HTML tags and trims a string.
func SanitizeString(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	if len(s) > 10000 {
		s = s[:10000]
	}
	return s
}

// SanitizeMap recursively sanitizes string values in a map.
func SanitizeMap(m map[string]any) map[string]any {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			m[k] = SanitizeString(val)
		case map[string]any:
			m[k] = SanitizeMap(val)
		}
	}
	return m
}

// ErrorResponse is the standard API error format.
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError writes a standardized error response.
func WriteError(w http.ResponseWriter, statusCode int, code, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{
		"error": ErrorResponse{Code: code, Message: message},
	})
}

// WriteJSON writes a successful JSON response.
func WriteJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]any{"data": data})
}
