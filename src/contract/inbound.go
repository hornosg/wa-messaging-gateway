// Package contract define los contratos cross-repo del proyecto (River job args).
//
// IMPORTANTE: InboundMessageArgs es el CONTRATO entre el messaging-gateway (productor)
// y el agent-runtime (consumidor, E04). Ambos repos deben usar el mismo `Kind()` y la
// misma forma JSON. Candidato a extraerse a un paquete compartido (wa-contracts o
// go-shared) y a formalizarse en el puerto AgentRuntime (D-03 / ADR-0003).
package contract

import "time"

// InboundMessageArgs — job de mensaje entrante encolado en River.
type InboundMessageArgs struct {
	TenantSlug        string    `json:"tenant_slug"`
	ProviderMessageID string    `json:"provider_message_id"` // id del mensaje en Kapso/Meta
	From              string    `json:"from"`
	To                string    `json:"to"`
	Text              string    `json:"text"`
	ReceivedAt        time.Time `json:"received_at"`
}

// Kind implementa river.JobArgs. Debe ser idéntico en el consumidor (E04).
func (InboundMessageArgs) Kind() string { return "inbound_message" }
