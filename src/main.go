// messaging-gateway (SVC-01) — receiver de webhooks de WhatsApp (Kapso):
// verifica firma, encola el mensaje en River y responde 200. Hexagonal (P-04),
// canonical logs (P-20). Ref: PROJECT.md, roadmap E03, ADR-0001.
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
	"github.com/hornosg/wa-messaging-gateway/src/inbound"
	"github.com/hornosg/wa-messaging-gateway/src/logging"
)

const serviceName = "messaging-gateway"

func main() {
	cfg := config.Load()
	log := logging.New(serviceName)
	ctx := context.Background()

	// Puerto de cola: River (real) o Log (dev).
	var queue inbound.MessageQueue
	var closeQueue func()
	switch cfg.QueueDriver {
	case "log":
		queue = inbound.NewLogQueue(log)
		log.Warn("startup.queue_driver", map[string]any{"driver": "log"})
	default:
		rq, closeFn, err := inbound.NewRiverQueue(ctx, cfg.RiverDSN())
		if err != nil {
			log.Error("startup.river_init_failed", map[string]any{"error": err.Error()})
			os.Exit(1)
		}
		queue, closeQueue = rq, closeFn
		log.Info("startup.queue_driver", map[string]any{"driver": "river"})
	}
	if closeQueue != nil {
		defer closeQueue()
	}

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

	// Shutdown ordenado.
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	log.Info("shutdown.start", nil)
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Info("shutdown.done", nil)
}
