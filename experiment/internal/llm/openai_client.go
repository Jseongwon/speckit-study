// internal/llm/openai_client.go
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

type OpenAIClient struct {
	Model  string
	apiKey string
	httpc  *http.Client
}

func NewOpenAIClient(model string) *OpenAIClient {
	return &OpenAIClient{
		Model:  model, // 예: "gpt-4.1"
		apiKey: os.Getenv("OPENAI_API_KEY"),
		httpc:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *OpenAIClient) Name() string { return c.Model }

func (c *OpenAIClient) Generate(ctx context.Context, prompt string) (string, error) {
	body := map[string]interface{}{
		"model": c.Model,
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
		"max_completion_tokens": 2048,
	}

	b, _ := json.Marshal(body)
	req, err := http.NewRequestWithContext(ctx, "POST",
		"https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	json.NewDecoder(resp.Body).Decode(&decoded)
	if len(decoded.Choices) == 0 {
		return "", errors.New("no choices found")
	}
	return decoded.Choices[0].Message.Content, nil
}
