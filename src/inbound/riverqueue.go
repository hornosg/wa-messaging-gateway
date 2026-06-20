package inbound

import (
	"context"

	"github.com/hornosg/wa-messaging-gateway/src/contract"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// RiverQueue — adaptador del puerto MessageQueue sobre River (STACK-04).
// Productor: solo inserta jobs; los workers los registra el agent-runtime (E04).
type RiverQueue struct {
	pool   *pgxpool.Pool
	client *river.Client[pgx.Tx]
}

// NewRiverQueue crea el pool, aplica las migraciones de River y devuelve el adaptador
// junto con una función de cierre.
func NewRiverQueue(ctx context.Context, dsn string) (*RiverQueue, func(), error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, err
	}

	driver := riverpgxv5.New(pool)

	migrator, err := rivermigrate.New(driver, nil)
	if err != nil {
		pool.Close()
		return nil, nil, err
	}
	if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		pool.Close()
		return nil, nil, err
	}

	client, err := river.NewClient(driver, &river.Config{})
	if err != nil {
		pool.Close()
		return nil, nil, err
	}

	return &RiverQueue{pool: pool, client: client}, func() { pool.Close() }, nil
}

func (q *RiverQueue) Enqueue(ctx context.Context, m InboundMessage) error {
	_, err := q.client.Insert(ctx, contract.InboundMessageArgs{
		TenantSlug:        m.TenantSlug,
		ProviderMessageID: m.ProviderMessageID,
		From:              m.From,
		To:                m.To,
		Text:              m.Text,
		ReceivedAt:        m.ReceivedAt,
	}, nil)
	return err
}
