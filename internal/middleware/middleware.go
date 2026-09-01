// Package middleware provides HTTP middleware utilities.
package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// maxSanitizeBodyBytes 与 httpapi.decode 的 1MiB 上限保持一致。
const maxSanitizeBodyBytes = 1 << 20

// MaxSanitizeFieldBytes 单字段最大长度（消毒后）。
// 超限不再静默截断（用户以为数据完整），而是整个请求 400 报错。
const MaxSanitizeFieldBytes = 10000

// SanitizeJSONBody 对任意字节流做 JSON 消毒（白名单富文本）：可解析为 JSON 时
// 递归消毒字符串值并重新序列化（password 保留）；不可解析时原样返回（nil, nil）。
// 该函数同时供 SanitizeBody 中间件与 httpapi.decode 使用——decode 侧兜底可关闭
// "Content-Type 声明为非 JSON 而 body 是 JSON" 的绕过路径（此前 text/plain 携带
// JSON 可跳过唯一一层服务端消毒，构成存储型 XSS 绕过）。
func SanitizeJSONBody(raw []byte) ([]byte, error) {
	if len(bytes.TrimSpace(raw)) == 0 {
		return raw, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber() // 保留大整数精度（budget_fen 等 int64 字段）
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, nil // 非 JSON（multipart 等）：原样处理，由 decode 层判定
	}
	cleanV, err := sanitizeValue(v)
	if err != nil {
		return nil, err
	}
	clean, err := json.Marshal(cleanV)
	if err != nil {
		return nil, err
	}
	return clean, nil
}

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
		clean, err := SanitizeJSONBody(raw)
		if err != nil {
			WriteError(w, http.StatusBadRequest, "FIELD_TOO_LONG", err.Error())
			return
		}
		if clean == nil {
			// 非法 JSON 原样放行，由下游 decode 层返回统一校验错误。
			r.Body = io.NopCloser(bytes.NewReader(raw))
			next.ServeHTTP(w, r)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(clean))
		next.ServeHTTP(w, r)
	})
}

// sanitizeValue recursively sanitizes string values inside arbitrary JSON.
// 字段超长（消毒后 > MaxSanitizeFieldBytes）导致整个请求 400（FIELD_TOO_LONG），
// 替代旧实现静默截断——用户提交超长内容时数据被悄悄改掉而不知情。
func sanitizeValue(v any) (any, error) {
	switch val := v.(type) {
	case string:
		return SanitizeStringStrict(val)
	case map[string]any:
		for k, item := range val {
			if strings.EqualFold(k, "password") {
				continue
			}
			nv, err := sanitizeValue(item)
			if err != nil {
				return nil, err
			}
			val[k] = nv
		}
		return val, nil
	case []any:
		for i := range val {
			nv, err := sanitizeValue(val[i])
			if err != nil {
				return nil, err
			}
			val[i] = nv
		}
		return val, nil
	default:
		return v, nil
	}
}

var htmlTagRe = regexp.MustCompile(`<[^>]*>`)

// 富文本白名单（内容字段允许保留的标签）；危险容器标签连同内容整体移除；
// a 仅保留 http/https/相对路径 href；img 仅保留 http/https/相对路径 src。其余标签剥掉但保留文本。
var (
	sanitizeAllowTags = map[string]bool{
		"p": true, "br": true, "strong": true, "b": true, "em": true, "i": true,
		"u": true, "s": true, "h1": true, "h2": true, "h3": true, "h4": true,
		"ul": true, "ol": true, "li": true, "blockquote": true, "span": true,
		"a": true, "img": true, "hr": true,
	}
	sanitizeDangerBlock = regexp.MustCompile(`(?is)<(script|style|iframe|object|embed|form|textarea|select|button)\b[^>]*>.*?<\s*/\s*(?:script|style|iframe|object|embed|form|textarea|select|button)\s*>`)
	sanitizeDangerSelf  = regexp.MustCompile(`(?is)<(script|style|iframe|object|embed|form|textarea|select|button)\b[^>]*?>`)
	sanitizeTag         = regexp.MustCompile(`(?is)<\s*(/?)([a-z0-9]+)([^>]*?)(\s*/?)\s*>`)
	sanitizeHref        = regexp.MustCompile(`(?i)^\s*(https?://|/|\./|#)`)
)

// sanitizeRichText 白名单消毒：危险标签连内容删；允许标签仅保留安全属性；其余剥标签保文本；
// 结尾做闭标签配对平衡（起始标签被剥时，孤立的闭合标签一并移除）。
func sanitizeRichText(s string) string {
	s = sanitizeDangerBlock.ReplaceAllString(s, "")
	s = sanitizeDangerSelf.ReplaceAllString(s, "")
	opened := map[string]int{}
	s = sanitizeTag.ReplaceAllStringFunc(s, func(m string) string {
		p := sanitizeTag.FindStringSubmatch(m)
		name := strings.ToLower(p[2])
		if !sanitizeAllowTags[name] {
			return "" // 剥标签，保留内部文本
		}
		closing := p[1] == "/"
		if closing {
			if opened[name] > 0 {
				opened[name]--
				return "</" + name + ">"
			}
			return "" // 孤立闭合标签（起始被剥/不匹配）
		}
		switch name {
		case "a":
			href := ""
			for _, attr := range strings.Fields(p[3]) {
				kv := strings.SplitN(attr, "=", 2)
				if len(kv) != 2 {
					continue
				}
				if !strings.EqualFold(strings.Trim(kv[0], `"' `), "href") {
					continue
				}
				v := strings.Trim(strings.Trim(kv[1], `"'`), " ")
				if sanitizeHref.MatchString(v) {
					href = v
				}
			}
			if href == "" {
				return "" // href 非法：整标签剥除（配对闭合由 opened 计数吸收）
			}
			opened[name]++
			return `<a href="` + href + `">`
		case "img":
			src := ""
			for _, attr := range strings.Fields(p[3]) {
				kv := strings.SplitN(attr, "=", 2)
				if len(kv) != 2 {
					continue
				}
				if !strings.EqualFold(strings.Trim(kv[0], `"' `), "src") {
					continue
				}
				v := strings.Trim(strings.Trim(kv[1], `"'`), " ")
				if sanitizeHref.MatchString(v) {
					src = v
				}
			}
			if src == "" {
				return ""
			}
			return `<img src="` + src + `">`
		case "br", "hr":
			return "<" + name + ">"
		default:
			opened[name]++
			return "<" + name + ">"
		}
	})
	return s
}

// SanitizeString removes HTML tags and trims a string (超长静默截断为
// MaxSanitizeFieldBytes，仅限文件路径等内部用途；HTTP 请求体请用
// SanitizeStringStrict——HTTP 字段超长不再静默截断）。
func SanitizeString(s string) string {
	s = htmlTagRe.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	if len(s) > MaxSanitizeFieldBytes {
		s = s[:MaxSanitizeFieldBytes]
	}
	return s
}

// SanitizeStringStrict like SanitizeString but errors when the sanitized value
// exceeds MaxSanitizeFieldBytes instead of silently truncating.
func SanitizeStringStrict(s string) (string, error) {
	s = sanitizeRichText(s)
	s = strings.TrimSpace(s)
	if len(s) > MaxSanitizeFieldBytes {
		return "", fmt.Errorf("field too long: %d chars (max %d)", len(s), MaxSanitizeFieldBytes)
	}
	return s, nil
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
