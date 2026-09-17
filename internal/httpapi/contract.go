package httpapi

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"drone-platform/internal/domain"
)

// dedup set for webhook event_id (in-memory; PG contract_events table is the canonical store)
// webhookDedup 已处理事件去重表（eventID -> 处理完成时间；24h 窗口，惰性清理）。
var webhookDedup sync.Map

// GET /api/v1/contract-templates
// 从 contract_templates 表读取；服务未装配或旧环境未跑种子迁移时兜底返回内置模板。
func (s *Server) listContractTemplates(w http.ResponseWriter, r *http.Request) {
	if s.contractTplSvc != nil {
		list, err := s.contractTplSvc.List(r.Context())
		if err != nil {
			slog.Error("list contract templates failed", "err", err)
			fail(w, r, http.StatusInternalServerError, errors.New("failed to load contract templates"))
			return
		}
		if len(list) > 0 {
			respond(w, r, http.StatusOK, list)
			return
		}
	}
	respond(w, r, http.StatusOK, domain.DefaultContractTemplates)
}

// POST /api/v1/admin/members/import 与 validAssociationRole 已移除。
//
// 它们服务于协会 8 级角色（association_members），而那套东西从未在生产使用：
// 表 0 行、前端零调用、8 个角色里只有 partner 参与过判定。
// 顺带记一笔：这个接口曾是**唯一**校验角色合法性的地方，
// 单条的 POST /api/v1/admin/association-members 从不校验，可写入任意字符串。
// POST /api/v1/assignments
func (s *Server) createAssignment(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	var in struct {
		OrderID  string `json:"order_id"`
		WorkerID string `json:"worker_id"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if in.OrderID == "" || in.WorkerID == "" {
		fail(w, r, http.StatusBadRequest, errors.New("order_id and worker_id are required"))
		return
	}
	asgn, err := s.labourSvc.CreateAssignment(r.Context(), a, in.OrderID, in.WorkerID)
	if err != nil {
		code := http.StatusForbidden
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	s.audit(r.Context(), a.ID, "assign_worker", "assignment", asgn.ID, in.OrderID)
	respond(w, r, http.StatusCreated, asgn)
}

// POST /api/v1/contracts/{id}/void
func (s *Server) voidContract(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if _, err := s.contracts.UpdateStatus(r.Context(), a, r.PathValue("id"), domain.ContractVoided); err != nil {
		code := http.StatusConflict
		if strings.Contains(err.Error(), "not found") {
			code = http.StatusNotFound
		}
		fail(w, r, code, err)
		return
	}
	s.audit(r.Context(), a.ID, "void_contract", "contract", r.PathValue("id"), "voided")
	respond(w, r, http.StatusOK, map[string]string{"status": "voided"})
}

// POST /api/v1/webhooks/signing
func (s *Server) signingWebhook(w http.ResponseWriter, r *http.Request) {
	var event struct {
		EventID    string `json:"event_id"`
		ContractID string `json:"contract_id"`
		Status     string `json:"status"`
		Signature  string `json:"signature"`
		Timestamp  int64  `json:"timestamp"`
	}
	if err := decode(r, &event); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}

	// P0 修复：签名校验不可缺省——此前 SIGNING_SECRET 未配置时跳过校验，
	// 任何人不签名即可伪造事件翻转合同状态。未配置一律拒绝；
	// 生产环境 config 硬校验已强制要求该密钥，此处兜底开发/误配场景。
	secret := os.Getenv("SIGNING_SECRET")
	if secret == "" {
		fail(w, r, http.StatusServiceUnavailable, errors.New("webhook signing is not configured (SIGNING_SECRET missing)"))
		return
	}
	// 时间戳新鲜度：拒绝重放旧事件（±5 分钟窗口）。
	if d := time.Now().Unix() - event.Timestamp; d > 300 || d < -300 {
		fail(w, r, http.StatusForbidden, errors.New("webhook timestamp expired"))
		return
	}
	ts := strconv.FormatInt(event.Timestamp, 10)
	if err := verifySigningSignature(secret, ts, event.EventID, event.ContractID, event.Status, event.Signature); err != nil {
		fail(w, r, http.StatusForbidden, fmt.Errorf("signature verification failed: %w", err))
		return
	}

	// P2 修复：去重必须原子占位——旧实现 Load→处理→Store 存在竞态，
	// 同一 event_id 的并发重试会双双通过去重并重复处理合同状态。
	// LoadOrStore 在处理前原子占位，同一 event_id 同时只有一个请求能进入处理流程。
	// 条目值存占位/完成时间：超过 24h 视为过期（惰性删除），map 内存有界。
	for {
		_, loaded := webhookDedup.LoadOrStore(event.EventID, time.Now())
		if !loaded {
			break // 本请求占位成功，进入处理
		}
		// 已被占用：24h 内视为重复；过期占位删除后重试重新占位（并发下仅一方能赢）。
		if t, ok := webhookDedup.Load(event.EventID); ok && time.Since(t.(time.Time)) < 24*time.Hour {
			respond(w, r, http.StatusOK, map[string]string{"received": event.EventID, "status": "duplicate"})
			return
		}
		webhookDedup.Delete(event.EventID)
	}

	newStatus, err := mapContractStatus(event.Status)
	if err != nil {
		// 参数错误：事件未消费，删除占位允许修正后重试。
		webhookDedup.Delete(event.EventID)
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if _, err := s.contracts.UpdateStatus(r.Context(), domain.Actor{ID: "system", Role: domain.RolePlatformAdmin}, event.ContractID, newStatus); err != nil {
		slog.Warn("signing webhook: failed to update contract status, not deduping", "contract_id", event.ContractID, "event_status", event.Status, "error", err)
		// 处理失败不消费事件：删除占位，发送方可安全重试（幂等语义）。
		webhookDedup.Delete(event.EventID)
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	webhookDedup.Store(event.EventID, time.Now()) // 处理完成时间
	s.audit(r.Context(), "system", "signing_callback", "contract", event.ContractID, event.Status)
	respond(w, r, http.StatusOK, map[string]string{"received": event.EventID})
}

// mapContractStatus maps external signing service status strings to internal contract status.
// 未知状态拒绝（返回错误）而非降级为 draft——此前未知事件会把已签合同改回草稿。
func mapContractStatus(eventStatus string) (domain.ContractStatus, error) {
	switch eventStatus {
	case "sent", "created":
		return domain.ContractSent, nil
	case "signing", "in_progress":
		return domain.ContractSigning, nil
	case "signed", "completed":
		return domain.ContractSigned, nil
	case "voided", "cancelled":
		return domain.ContractVoided, nil
	case "expired":
		return domain.ContractExpired, nil
	default:
		return "", fmt.Errorf("unknown webhook event status %q", eventStatus)
	}
}

func verifySigningSignature(secret, timestamp, eventID, contractID, status, signature string) error {
	mac := hmac.New(sha256.New, []byte(secret))
	payload := timestamp + "." + eventID + "." + contractID + "." + status
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return errors.New("signature mismatch")
	}
	return nil
}
