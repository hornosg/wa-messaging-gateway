// Package outbound — envío de respuestas al cliente vía Kapso (E03).
package outbound

import (
	"context"

	"github.com/hornosg/wa-messaging-gateway/src/logging"
)

// OutboundMessage — respuesta a enviar a un contacto.
type OutboundMessage struct {
	TenantSlug string
	To         string
	Text       string
	Handoff    bool
}

// Sender — puerto de salida: envía el mensaje al BSP. Adaptadores: Kapso / Stub.
type Sender interface {
	Send(ctx context.Context, m OutboundMessage) error
}

// StubSender — dev/sin key: loguea en vez de enviar. El real es KapsoSender.
type StubSender struct{ log *logging.Logger }

func NewStubSender(log *logging.Logger) *StubSender { return &StubSender{log: log} }

func (s *StubSender) Send(_ context.Context, m OutboundMessage) error {
	s.log.Info("kapso.send_stub", map[string]any{
		"to": m.To, "tenant_slug": m.TenantSlug, "handoff": m.Handoff, "text": m.Text,
		"note": "KAPSO_DRIVER=stub o sin KAPSO_API_KEY — no se envió a Kapso",
	})
	return nil
}
