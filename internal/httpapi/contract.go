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

// POST /api/v1/admin/members/import
// 同步批量导入：{"members":[{"user_id","enterprise_id","role"}]}，逐条落库。
// role 为空默认 member；行级校验失败不影响其余行，返回导入数与逐行失败明细。
func (s *Server) importMembers(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	var in struct {
		Members []struct {
			UserID       string `json:"user_id"`
			EnterpriseID string `json:"enterprise_id"`
			Role         string `json:"role"`
		} `json:"members"`
	}
	if err := decode(r, &in); err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if len(in.Members) == 0 {
		fail(w, r, http.StatusBadRequest, errors.New("members is required"))
		return
	}
	if s.assocMemberSvc == nil {
		fail(w, r, http.StatusInternalServerError, errors.New("member service unavailable"))
		return
	}

	type failedRow struct {
		Index  int    `json:"index"`
		UserID string `json:"user_id"`
		Error  string `json:"error"`
	}
	imported := 0
	var failed []failedRow
	for i, m := range in.Members {
		if m.UserID == "" {
			failed = append(failed, failedRow{Index: i, Error: "user_id is required"})
			continue
		}
		role := domain.AssociationRole(m.Role)
		if m.Role == "" {
			role = domain.AssocMember
		}
		if !validAssociationRole(role) {
			failed = append(failed, failedRow{Index: i, UserID: m.UserID, Error: "invalid role: " + m.Role})
			continue
		}
		if _, err := s.assocMemberSvc.AddMember(r.Context(), m.UserID, m.EnterpriseID, role); err != nil {
			failed = append(failed, failedRow{Index: i, UserID: m.UserID, Error: err.Error()})
			continue
		}
		imported++
	}
	s.audit(r.Context(), a.ID, "import_members", "association_member", "", fmt.Sprintf("imported=%d failed=%d", imported, len(failed)))
	respond(w, r, http.StatusOK, map[string]any{
		"imported": imported,
		"failed":   failed,
		"total":    len(in.Members),
	})
}

func validAssociationRole(role domain.AssociationRole) bool {
	switch role {
	case domain.AssocPresident, domain.AssocVicePresident, domain.AssocSecretary,
		domain.AssocDeptHead, domain.AssocMember, domain.AssocPartner,
		domain.AssocCollege, domain.AssocGuest:
		return true
	}
	return false
}

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

	// Check deduplication. Process first, then mark as done — if processing fails
	// the event_id is not consumed and the sender can safely retry.
	if _, loaded := webhookDedup.Load(event.EventID); loaded {
		respond(w, r, http.StatusOK, map[string]string{"received": event.EventID, "status": "duplicate"})
		return
	}

	newStatus, err := mapContractStatus(event.Status)
	if err != nil {
		fail(w, r, http.StatusBadRequest, err)
		return
	}
	if _, err := s.contracts.UpdateStatus(r.Context(), domain.Actor{ID: "system", Role: domain.RolePlatformAdmin}, event.ContractID, newStatus); err != nil {
		slog.Warn("signing webhook: failed to update contract status, not deduping", "contract_id", event.ContractID, "event_status", event.Status, "error", err)
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	webhookDedup.Store(event.EventID, true)
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
