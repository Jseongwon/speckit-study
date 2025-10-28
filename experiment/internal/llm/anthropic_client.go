// internal/llm/anthropic_client.go
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"
)

type AnthropicClient struct {
	Model  string
	apiKey string
	client *http.Client
}

func NewAnthropicClient(model string) *AnthropicClient {
	return &AnthropicClient{
		Model:  model, // 예: "claude-3.5-sonnet-4.5"
		apiKey: os.Getenv("ANTHROPIC_API_KEY"),
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *AnthropicClient) Name() string { return c.Model }

func (c *AnthropicClient) Generate(ctx context.Context, prompt string) (string, error) {
	reqBody := map[string]interface{}{
		"model":      c.Model,
		"max_tokens": 2048,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	b, _ := json.Marshal(reqBody)
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.anthropic.com/v1/messages", bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", c.apiKey)
	// Anthropic API는 `anthropic-version` 헤더로 버전을 명시하도록 요구합니다. :contentReference[oaicite:13]{index=13}
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var decoded struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	json.NewDecoder(resp.Body).Decode(&decoded)

	if len(decoded.Content) == 0 {
		return "", errors.New("no content found")
	}
	return decoded.Content[0].Text, nil
}
