// Package feishu provides a Feishu (飞书) webhook notifier.
package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/EthanShen10086/voxera-kit/notification"
)

// Adapter implements notification.Notifier for Feishu group robot webhooks.
type Adapter struct {
	webhookURL string
	client     *http.Client
}

// New creates a new Feishu notifier.
func New(cfg notification.Config) *Adapter {
	return &Adapter{
		webhookURL: cfg.WebhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Send delivers a message via Feishu webhook using interactive card format.
func (a *Adapter) Send(ctx context.Context, msg *notification.Message) (*notification.DeliveryResult, error) {
	payload := map[string]any{
		"msg_type": "interactive",
		"card": map[string]any{
			"header": map[string]any{
				"title": map[string]string{
					"tag":     "plain_text",
					"content": msg.Title,
				},
			},
			"elements": []map[string]string{
				{
					"tag":     "markdown",
					"content": msg.Content,
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return failResult(err), nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.webhookURL, bytes.NewReader(body))
	if err != nil {
		return failResult(err), nil
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return failResult(err), nil
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return failResult(fmt.Errorf("feishu: status %d, body: %s", resp.StatusCode, respBody)), nil
	}

	return &notification.DeliveryResult{
		Status:    notification.StatusDelivered,
		Channel:   notification.ChannelFeishu,
		Timestamp: time.Now(),
	}, nil
}

// Channel returns the channel type.
func (a *Adapter) Channel() notification.ChannelType {
	return notification.ChannelFeishu
}

func failResult(err error) *notification.DeliveryResult {
	return &notification.DeliveryResult{
		Status:    notification.StatusFailed,
		Channel:   notification.ChannelFeishu,
		Timestamp: time.Now(),
		Error:     err.Error(),
	}
}
