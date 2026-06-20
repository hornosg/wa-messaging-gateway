package config

import (
	"fmt"
	"os"
)

// Config del messaging-gateway (SVC-01). Todo por entorno (RULE-02, 12-factor).
type Config struct {
	Port               string
	DBHost             string
	DBPort             string
	DBName             string
	DBUser             string
	DBPassword         string
	DBSSLMode          string
	KapsoWebhookSecret string // RULE-05: verificación de firma del webhook
	QueueDriver        string // "river" (real) | "log" (dev: solo loguea el encolado)
}

func Load() Config {
	return Config{
		Port:               env("PORT", "8101"),
		DBHost:             env("DB_HOST", "lab-postgres"),
		DBPort:             env("DB_PORT", "5432"),
		DBName:             env("DB_NAME", "whatsapp_agent"),
		DBUser:             env("DB_USER", "whatsapp_agent"),
		DBPassword:         env("DB_PASSWORD", "whatsapp_agent"),
		DBSSLMode:          env("DB_SSLMODE", "disable"),
		KapsoWebhookSecret: env("KAPSO_WEBHOOK_SECRET", ""),
		QueueDriver:        env("QUEUE_DRIVER", "river"),
	}
}

// RiverDSN apunta al schema `river` (la cola vive ahí, ver db/init de services).
func (c Config) RiverDSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s&search_path=river",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
