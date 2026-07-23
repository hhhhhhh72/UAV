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
func (s *Server) listContractTemplates(w http.ResponseWriter, r *http.Request) {
	respond(w, r, http.StatusOK, []map[string]string{
		{"id": "tpl-001", "name": "标准无人机服务合同", "version": "1"},
		{"id": "tpl-002", "name": "无人机买卖协议", "version": "1"},
	})
}

// POST /api/v1/admin/members/import
func (s *Server) importMembers(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	if a.Role != domain.RoleAssociationAdmin && a.Role != domain.RolePlatformAdmin {
		fail(w, r, http.StatusForbidden, errors.New("admin permission required"))
		return
	}
	respond(w, r, http.StatusAccepted, map[string]string{
		"task_id": fmt.Sprintf("import-%d", time.Now().UnixNano()),
		"status":  "processing",
	})
}

// POST /api/v1/assignments
func (s *Server) createAssignment(w http.ResponseWriter, r *http.Request) {
	_, ok := authenticatedActor(r)
	if !ok { fail(w, r, http.StatusUnauthorized, errors.New("authentication required")); return }
	var in struct {
		OrderID  string `json:"order_id"`
		WorkerID string `json:"worker_id"`
	}
	if err := decode(r, &in); err != nil { fail(w, r, http.StatusBadRequest, err); return }
	now := time.Now()
	asgn := domain.Assignment{ID: fmt.Sprintf("assign-%d", now.UnixNano()), OrderID: in.OrderID, WorkerID: in.WorkerID, Status: "assigned", CreatedAt: now}
	respond(w, r, http.StatusCreated, asgn)
}

// POST /api/v1/contracts/{id}/void
func (s *Server) voidContract(w http.ResponseWriter, r *http.Request) {
	a, ok := authenticatedActor(r)
	if !ok {
		fail(w, r, http.StatusUnauthorized, errors.New("authentication required"))
		return
	}
	if _, err := s.contracts.UpdateStatus(a, r.PathValue("id"), domain.ContractVoided); err != nil {
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

	// Verify signature if SIGNING_SECRET is configured.
	if secret := os.Getenv("SIGNING_SECRET"); secret != "" {
		ts := strconv.FormatInt(event.Timestamp, 10)
		if err := verifySigningSignature(secret, ts, event.EventID, event.ContractID, event.Status, event.Signature); err != nil {
			fail(w, r, http.StatusForbidden, fmt.Errorf("signature verification failed: %w", err))
			return
		}
	}

	// Check deduplication. Process first, then mark as done — if processing fails
	// the event_id is not consumed and the sender can safely retry.
	if _, loaded := webhookDedup.Load(event.EventID); loaded {
		respond(w, r, http.StatusOK, map[string]string{"received": event.EventID, "status": "duplicate"})
		return
	}

	newStatus := mapContractStatus(event.Status)
	if _, err := s.contracts.UpdateStatus(domain.Actor{ID: "system", Role: domain.RolePlatformAdmin}, event.ContractID, newStatus); err != nil {
		slog.Warn("signing webhook: failed to update contract status, not deduping", "contract_id", event.ContractID, "event_status", event.Status, "error", err)
		fail(w, r, http.StatusInternalServerError, err)
		return
	}
	webhookDedup.Store(event.EventID, true)
	s.audit(r.Context(), "system", "signing_callback", "contract", event.ContractID, event.Status)
	respond(w, r, http.StatusOK, map[string]string{"received": event.EventID})
}

// mapContractStatus maps external signing service status strings to internal contract status.
func mapContractStatus(eventStatus string) domain.ContractStatus {
	switch eventStatus {
	case "sent", "created":
		return domain.ContractSent
	case "signing", "in_progress":
		return domain.ContractSigning
	case "signed", "completed":
		return domain.ContractSigned
	case "voided", "cancelled":
		return domain.ContractVoided
	case "expired":
		return domain.ContractExpired
	default:
		slog.Warn("signing webhook: unknown contract event status, defaulting to draft", "event_status", eventStatus)
		return domain.ContractDraft
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
