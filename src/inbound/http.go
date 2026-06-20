package inbound

import (
	"io"
	"net/http"

	"github.com/hornosg/wa-messaging-gateway/src/logging"
)

// Handler expone los endpoints HTTP del gateway.
type Handler struct {
	svc            *Service
	webhookSecret  string
	log            *logging.Logger
	verifyDisabled bool // true si no hay secret (dev): se loguea warn y se acepta
}

func NewHandler(svc *Service, webhookSecret string, log *logging.Logger) *Handler {
	return &Handler{
		svc:            svc,
		webhookSecret:  webhookSecret,
		log:            log,
		verifyDisabled: webhookSecret == "",
	}
}

func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", h.health)
	mux.HandleFunc("POST /api/v1/webhooks/kapso", h.kapsoWebhook)
	return mux
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) kapsoWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MiB
	if err != nil {
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}

	// RULE-05: verificar firma antes de procesar.
	if h.verifyDisabled {
		h.log.Warn("webhook.signature_unverified", map[string]any{"note": "KAPSO_WEBHOOK_SECRET vacío (solo dev)"})
	} else if !VerifySignature(h.webhookSecret, body, r.Header.Get("X-Kapso-Signature")) {
		h.log.Warn("webhook.signature_invalid", nil)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	msg, err := ParseWebhook(body)
	if err != nil {
		h.log.Warn("webhook.bad_payload", map[string]any{"error": err.Error()})
		http.Error(w, "bad payload", http.StatusBadRequest)
		return
	}

	// Encolar ANTES de responder 200: si falla, 500 para que Kapso reintente.
	if err := h.svc.Handle(r.Context(), msg); err != nil {
		http.Error(w, "enqueue failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"accepted"}`))
}
