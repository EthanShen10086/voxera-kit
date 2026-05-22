// Package wecom provides a WeChat Work (企业微信) webhook notifier.
package wecom

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

// Adapter implements notification.Notifier for WeChat Work group robot webhooks.
type Adapter struct {
	webhookURL string
	client     *http.Client
}

// New creates a new WeChat Work notifier.
func New(cfg notification.Config) *Adapter {
	return &Adapter{
		webhookURL: cfg.WebhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

// Send delivers a message via WeChat Work webhook.
func (a *Adapter) Send(ctx context.Context, msg *notification.Message) (*notification.DeliveryResult, error) {
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": formatContent(msg),
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
		return failResult(fmt.Errorf("wecom: status %d, body: %s", resp.StatusCode, respBody)), nil
	}

	return &notification.DeliveryResult{
		Status:    notification.StatusDelivered,
		Channel:   notification.ChannelWecom,
		Timestamp: time.Now(),
	}, nil
}

// Channel returns the channel type.
func (a *Adapter) Channel() notification.ChannelType {
	return notification.ChannelWecom
}

func formatContent(msg *notification.Message) string {
	if msg.Title != "" {
		return fmt.Sprintf("## %s\n\n%s", msg.Title, msg.Content)
	}
	return msg.Content
}

func failResult(err error) *notification.DeliveryResult {
	return &notification.DeliveryResult{
		Status:    notification.StatusFailed,
		Channel:   notification.ChannelWecom,
		Timestamp: time.Now(),
		Error:     err.Error(),
	}
}
