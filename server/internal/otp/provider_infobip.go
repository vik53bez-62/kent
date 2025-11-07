package otp

import (
  "bytes"
  "context"
  "encoding/json"
  "fmt"
  "net/http"
  "time"
)

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
  payload := map[string]any{
    "messages": []map[string]any{
      {
        "from": i.From,
        "destinations": []map[string]string{{"to": to}},
        "text": text,
      },
    },
  }
  b, _ := json.Marshal(payload)
  req, _ := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/sms/2/text/advanced", i.BaseURL), bytes.NewReader(b))
  req.Header.Set("Authorization", "App "+i.APIKey)
  req.Header.Set("Content-Type", "application/json")
  resp, err := i.Client.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()
  if resp.StatusCode >= 300 {
    return fmt.Errorf("infobip sms failed: status=%d", resp.StatusCode)
  }
  return nil
}
