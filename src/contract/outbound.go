package contract

// OutboundMessageArgs — job de respuesta saliente. El agent-runtime lo encola
// (cola "outbound") y el messaging-gateway lo consume para enviarlo a Kapso.
// Contrato cross-repo (ADR-0003): mismo Kind + JSON en ambos repos.
type OutboundMessageArgs struct {
	TenantSlug string `json:"tenant_slug"`
	To         string `json:"to"`
	Text       string `json:"text"`
	Handoff    bool   `json:"handoff"`
}

func (OutboundMessageArgs) Kind() string { return "outbound_message" }

// Nombres de cola River (separan productor/consumidor entre servicios).
const (
	QueueInbound  = "inbound"
	QueueOutbound = "outbound"
)
