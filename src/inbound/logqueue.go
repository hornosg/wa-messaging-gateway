package inbound

import (
	"context"

	"github.com/hornosg/wa-messaging-gateway/src/logging"
)

// LogQueue — adaptador de dev del puerto MessageQueue: no encola en River, solo
// loguea. Útil para correr el gateway sin tocar la cola (QUEUE_DRIVER=log).
type LogQueue struct {
	log *logging.Logger
}

func NewLogQueue(log *logging.Logger) *LogQueue { return &LogQueue{log: log} }

func (q *LogQueue) Enqueue(_ context.Context, m InboundMessage) error {
	q.log.Warn("inbound.enqueue_logonly", map[string]any{
		"tenant_slug": m.TenantSlug, "provider_message_id": m.ProviderMessageID,
		"from": m.From, "note": "QUEUE_DRIVER=log — no se encoló en River",
	})
	return nil
}
