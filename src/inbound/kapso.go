package inbound

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// VerifySignature valida la firma HMAC-SHA256 del webhook (RULE-05).
// Esquema asumido: header `sha256=<hex(hmac_sha256(body, secret))>`.
// TODO(kapso): confirmar nombre de header y esquema exacto con la doc de Kapso.
func VerifySignature(secret string, body []byte, signatureHeader string) bool {
	if secret == "" || signatureHeader == "" {
		return false
	}
	want := strings.TrimPrefix(signatureHeader, "sha256=")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	got := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(got), []byte(want))
}

// kapsoWebhook — forma permisiva del payload entrante.
// TODO(kapso): mapear a la forma real del webhook de Kapso/Meta.
type kapsoWebhook struct {
	TenantSlug string `json:"tenant_slug"`
	MessageID  string `json:"message_id"`
	From       string `json:"from"`
	To         string `json:"to"`
	Text       string `json:"text"`
}

var ErrEmptyPayload = errors.New("payload vacío o sin texto")

// ParseWebhook normaliza el payload a un InboundMessage de dominio.
func ParseWebhook(body []byte) (InboundMessage, error) {
	var w kapsoWebhook
	if err := json.Unmarshal(body, &w); err != nil {
		return InboundMessage{}, err
	}
	if w.MessageID == "" || w.From == "" {
		return InboundMessage{}, ErrEmptyPayload
	}
	return InboundMessage{
		TenantSlug:        w.TenantSlug,
		ProviderMessageID: w.MessageID,
		From:              w.From,
		To:                w.To,
		Text:              w.Text,
		ReceivedAt:        time.Now().UTC(),
	}, nil
}
