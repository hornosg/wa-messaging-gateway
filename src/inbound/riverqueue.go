package inbound

import (
	"context"

	"github.com/hornosg/wa-messaging-gateway/src/contract"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

// RiverQueue — adaptador del puerto MessageQueue: inserta el job inbound en River
// (cola "inbound"). El cliente River lo crea y administra main (también consume outbound).
type RiverQueue struct {
	client *river.Client[pgx.Tx]
}

func NewRiverQueue(client *river.Client[pgx.Tx]) *RiverQueue {
	return &RiverQueue{client: client}
}

func (q *RiverQueue) Enqueue(ctx context.Context, m InboundMessage) error {
	_, err := q.client.Insert(ctx, contract.InboundMessageArgs{
		TenantSlug:        m.TenantSlug,
		ProviderMessageID: m.ProviderMessageID,
		From:              m.From,
		To:                m.To,
		Text:              m.Text,
		ReceivedAt:        m.ReceivedAt,
	}, &river.InsertOpts{Queue: contract.QueueInbound})
	return err
}
