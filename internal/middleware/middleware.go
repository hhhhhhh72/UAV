// Package middleware provides HTTP middleware utilities.
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// maxSanitizeBodyBytes 与 httpapi.decode 的 1MiB 上限保持一致。
const maxSanitizeBodyBytes = 1 << 20

// SanitizeBody reads JSON request bodies and recursively strips HTML tags from
// string values (defense-in-depth against stored XSS), then re-injects the
// sanitized body for downstream handlers. Non-JSON content types (multipart
// uploads 等) 与只读方法原样放行；非法 JSON 也原样放行，由 decode 层返回统一
// 校验错误。字段名 password 不做消毒，保证登录/注册凭据保真。
func SanitizeBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}
		if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
			next.ServeHTTP(w, r)
			return
		}
		if r.Body == nil {
			next.ServeHTTP(w, r)
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, maxSanitizeBodyBytes+1))
		_ = r.Body.Close()
		if err != nil {
			WriteError(w, http.StatusBadRequest, "BAD_REQUEST", "read request body failed")
			return
		}
		if len(raw) > maxSanitizeBodyBytes {
			WriteError(w, http.StatusRequestEntityTooLarge, "PAYLOAD_TOO_LARGE", "request body too large")
			return
		}
		if len(bytes.TrimSpace(raw)) == 0 {
			r.Body = io.NopCloser(bytes.NewReader(raw))
			next.ServeHTTP(w, r)
			return
		}
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber() // 保留大整数精度（budget_fen 等 int64 字段）
		var v any
		if err := dec.Decode(&v); err != nil {
			// 非法 JSON 原样放行，由下游 decode 层返回统一校验错误。
			r.Body = io.NopCloser(bytes.NewReader(raw))
			next.ServeHTTP(w, r)
			return
		}
		clean, err := json.Marshal(sanitizeValue(v))
		if err != nil {
			r.Body = io.NopCloser(bytes.NewReader(raw))
			next.ServeHTTP(w, r)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(clean))
		next.ServeHTTP(w, r)
	})
}

// sanitizeValue recursively sanitizes string values inside arbitrary JSON.
func sanitizeValue(v any) any {
	switch val := v.(type) {
	case string:
		return SanitizeString(val)
	case map[string]any:
		for k, item := range val {
			if strings.EqualFold(k, "password") {
				continue
			}
			val[k] = sanitizeValue(item)
		}
		return val
	case []any:
		for i := range val {
			val[i] = sanitizeValue(val[i])
		}
		return val
	default:
		return v
	}
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
