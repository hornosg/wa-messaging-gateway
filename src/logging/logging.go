// Package logging envuelve el canonical logger genérico del lab (go-shared,
// envelope ADR-001) en una API mínima. P-20: canonical logs desde el día uno.
package logging

import gs "github.com/hornosg/go-shared/infrastructure/logging"

type Logger struct {
	c *gs.CanonicalLogger
}

func New(service string) *Logger {
	return &Logger{c: gs.NewCanonicalLogger(service)}
}

func (l *Logger) Info(event string, fields map[string]any)  { l.c.Emit("info", event, fields) }
func (l *Logger) Warn(event string, fields map[string]any)  { l.c.Emit("warn", event, fields) }
func (l *Logger) Error(event string, fields map[string]any) { l.c.Emit("error", event, fields) }
