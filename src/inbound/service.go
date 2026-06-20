package inbound

import (
	"context"

	"github.com/hornosg/wa-messaging-gateway/src/logging"
)

// Service — caso de uso de ingreso: recibe un InboundMessage ya verificado y lo
// encola. Responder 200 al BSP recién después de encolar (no perder el mensaje).
type Service struct {
	queue MessageQueue
	log   *logging.Logger
}

func NewService(queue MessageQueue, log *logging.Logger) *Service {
	return &Service{queue: queue, log: log}
}

func (s *Service) Handle(ctx context.Context, m InboundMessage) error {
	if err := s.queue.Enqueue(ctx, m); err != nil {
		s.log.Error("inbound.enqueue_failed", map[string]any{
			"tenant_slug": m.TenantSlug, "provider_message_id": m.ProviderMessageID, "error": err.Error(),
		})
		return err
	}
	s.log.Info("inbound.enqueued", map[string]any{
		"tenant_slug": m.TenantSlug, "provider_message_id": m.ProviderMessageID,
	})
	return nil
}
