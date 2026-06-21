package outbound

import (
	"context"

	"github.com/hornosg/wa-messaging-gateway/src/contract"
	"github.com/riverqueue/river"
)

// Worker — consume jobs "outbound_message" (cola "outbound") y los envía vía Sender.
type Worker struct {
	river.WorkerDefaults[contract.OutboundMessageArgs]
	sender Sender
}

func NewWorker(sender Sender) *Worker { return &Worker{sender: sender} }

func (w *Worker) Work(ctx context.Context, job *river.Job[contract.OutboundMessageArgs]) error {
	a := job.Args
	return w.sender.Send(ctx, OutboundMessage{
		TenantSlug: a.TenantSlug,
		To:         a.To,
		Text:       a.Text,
		Handoff:    a.Handoff,
	})
}
