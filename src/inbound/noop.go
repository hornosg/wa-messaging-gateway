package inbound

import (
	"context"

	"github.com/hornosg/wa-messaging-gateway/src/contract"
	"github.com/riverqueue/river"
)

// NoopInboundWorker registra el kind "inbound_message" en el bundle del cliente
// para que River permita Insert (lo valida). NUNCA corre: este servicio solo
// trabaja la cola "outbound".
type NoopInboundWorker struct {
	river.WorkerDefaults[contract.InboundMessageArgs]
}

func NewNoopInboundWorker() *NoopInboundWorker { return &NoopInboundWorker{} }

func (*NoopInboundWorker) Work(context.Context, *river.Job[contract.InboundMessageArgs]) error {
	return nil
}
