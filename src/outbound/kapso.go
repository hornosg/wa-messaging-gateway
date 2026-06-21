package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// KapsoSender — envío real a la API de Kapso (necesita KAPSO_API_KEY).
// TODO(kapso): confirmar endpoint/payload exactos con la doc de Kapso. La forma
// de abajo es un placeholder; se ajusta al cablear la cuenta.
type KapsoSender struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

func NewKapsoSender(apiKey, baseURL string) *KapsoSender {
	if baseURL == "" {
		baseURL = "https://api.kapso.ai" // TODO: confirmar
	}
	return &KapsoSender{apiKey: apiKey, baseURL: baseURL, http: &http.Client{Timeout: 15 * time.Second}}
}

type kapsoSendReq struct {
	To   string `json:"to"`
	Text string `json:"text"`
}

func (s *KapsoSender) Send(ctx context.Context, m OutboundMessage) error {
	body, _ := json.Marshal(kapsoSendReq{To: m.To, Text: m.Text})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/v1/messages", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("kapso send %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
