package inbound

import (
	"context"
	"time"
)

// InboundMessage — mensaje entrante de un contacto (G-02), normalizado desde el
// webhook del BSP (Kapso). VO de dominio, independiente del proveedor.
type InboundMessage struct {
	TenantSlug        string
	ProviderMessageID string
	From              string
	To                string
	Text              string
	ReceivedAt        time.Time
}

// MessageQueue — puerto de salida: el dominio encola el mensaje para que el
// Agent Runtime lo procese async. Adaptadores: River (real) / Log (dev).
type MessageQueue interface {
	Enqueue(ctx context.Context, m InboundMessage) error
}
