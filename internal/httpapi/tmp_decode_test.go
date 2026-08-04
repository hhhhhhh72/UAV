package httpapi

import (
	"encoding/json"
	"testing"

	"drone-platform/internal/domain"
)

func TestTmpDecodeAttachment(t *testing.T) {
	var in struct {
		Attachments []domain.Attachment `json:"attachments"`
	}
	err := json.Unmarshal([]byte(`{"attachments":[{"name":"doc.pdf","size":"1MB","url":"/u/a.pdf"}]}`), &in)
	t.Logf("err=%v attachments=%+v", err, in.Attachments)
}
