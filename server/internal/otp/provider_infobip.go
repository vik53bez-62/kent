package otp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

var ErrProviderNotConfigured = errors.New("infobip: base URL or API key missing")

type SMSProvider interface {
	SendSMS(ctx context.Context, to, text string) error
}

type Infobip struct {
	BaseURL string
	APIKey  string
	From    string
	Client  *http.Client
}

func NewInfobip(baseURL, apiKey, from string) *Infobip {
	return &Infobip{
		BaseURL: baseURL,
		APIKey:  apiKey,
		From:    from,
		Client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (i *Infobip) SendSMS(ctx context.Context, to, text string) error {
	if i.BaseURL == "" || i.APIKey == "" {
		return ErrProviderNotConfigured
	}

	payload := map[string]any{
		"messages": []map[string]any{
			{
				"from":         i.From,
				"destinations": []map[string]string{{"to": to}},
				"text":         text,
			},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/sms/2/text/advanced", i.BaseURL), bytes.NewReader(body))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "App "+i.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := i.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("infobip sms failed: status=%d", resp.StatusCode)
	}

	return nil
}
