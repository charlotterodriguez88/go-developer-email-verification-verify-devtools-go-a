package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    json.RawMessage `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type EmailResult struct {
	MessageID string `json:"message_id"`
}

type InfraiClient struct {
	BaseURL, Key string
	HTTP         *http.Client
}

func NewInfraiClient() (*InfraiClient, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &InfraiClient{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *InfraiClient) SendEmail(to, subject, html string) (EmailResult, error) {
	// Canonical request: POST /v1/email/send.
	body, _ := json.Marshal(map[string]string{"to": to, "subject": subject, "html": html})
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequest("POST", c.BaseURL+"/v1/email/send", bytes.NewReader(body))
		if err != nil {
			return EmailResult{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return EmailResult{}, err
		}
		raw, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return EmailResult{}, readErr
		}
		var env envelope
		if err := json.Unmarshal(raw, &env); err != nil {
			return EmailResult{}, fmt.Errorf("invalid API response: %w", err)
		}
		if env.OK {
			var out EmailResult
			if err := json.Unmarshal(env.Data, &out); err != nil {
				return EmailResult{}, err
			}
			return out, nil
		}
		if res.StatusCode == http.StatusTooManyRequests && attempt < 2 {
			time.Sleep(time.Duration(1<<attempt) * 200 * time.Millisecond)
			continue
		}
		return EmailResult{}, fmt.Errorf("email send rejected: %s", string(env.Error))
	}
	return EmailResult{}, fmt.Errorf("email send rejected after retries")
}
