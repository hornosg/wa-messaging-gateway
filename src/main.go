// messaging-gateway (SVC-01) — receiver de webhooks de WhatsApp (Kapso):
// verifica firma, encola el inbound en River y responde 200. Además CONSUME la
// cola "outbound" (respuestas del agent-runtime) y las envía a Kapso. Hexagonal
// (P-04), canonical logs (P-20). Ref: PROJECT.md, roadmap E03, ADR-0001/0003.
package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hornosg/wa-messaging-gateway/src/config"
	"github.com/hornosg/wa-messaging-gateway/src/contract"
	"github.com/hornosg/wa-messaging-gateway/src/inbound"
	"github.com/hornosg/wa-messaging-gateway/src/logging"
	"github.com/hornosg/wa-messaging-gateway/src/outbound"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

const serviceName = "messaging-gateway"

func main() {
	cfg := config.Load()
	log := logging.New(serviceName)
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.RiverDSN())
	if err != nil {
		log.Error("startup.db_pool_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	defer pool.Close()
	driver := riverpgxv5.New(pool)

	if migrator, err := rivermigrate.New(driver, nil); err == nil {
		if _, err := migrator.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
			log.Error("startup.river_migrate_failed", map[string]any{"error": err.Error()})
			os.Exit(1)
		}
	}

	// Sender de salida: Kapso (real) o Stub (dev/sin key).
	var sender outbound.Sender
	if cfg.KapsoDriver == "stub" || cfg.KapsoAPIKey == "" {
		sender = outbound.NewStubSender(log)
		log.Warn("startup.kapso_driver", map[string]any{"driver": "stub", "reason": "KAPSO_DRIVER=stub o KAPSO_API_KEY vacío"})
	} else {
		sender = outbound.NewKapsoSender(cfg.KapsoAPIKey, cfg.KapsoBaseURL)
		log.Info("startup.kapso_driver", map[string]any{"driver": "kapso"})
	}

	// Cliente River: consume la cola "outbound" + sirve para insertar "inbound".
	workers := river.NewWorkers()
	river.AddWorker(workers, outbound.NewWorker(sender))     // real: consume "outbound"
	river.AddWorker(workers, inbound.NewNoopInboundWorker()) // solo para validar Insert("inbound")
	client, err := river.NewClient(driver, &river.Config{
		Queues:  map[string]river.QueueConfig{contract.QueueOutbound: {MaxWorkers: cfg.MaxWorkers}},
		Workers: workers,
	})
	if err != nil {
		log.Error("startup.river_client_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	if err := client.Start(ctx); err != nil {
		log.Error("startup.river_start_failed", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	log.Info("startup.consuming", map[string]any{"queue": contract.QueueOutbound})

	queue := inbound.NewRiverQueue(client)
	svc := inbound.NewService(queue, log)
	handler := inbound.NewHandler(svc, cfg.KapsoWebhookSecret, log)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
	}
	go func() {
		log.Info("startup.listening", map[string]any{"port": cfg.Port})
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server.failed", map[string]any{"error": err.Error()})
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutdown.start", nil)
	shutdownCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	_ = client.Stop(shutdownCtx)
	log.Info("shutdown.done", nil)
}
