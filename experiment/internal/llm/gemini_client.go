// internal/llm/gemini_client.go
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"
)

type GeminiClient struct {
	Model  string
	apiKey string
	httpc  *http.Client
}

func NewGeminiClient(model string) *GeminiClient {
	return &GeminiClient{
		Model:  model, // 예: "gemini-2.5-pro" 또는 "gemini-2.5-flash"
		apiKey: os.Getenv("GEMINI_API_KEY"),
		httpc:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *GeminiClient) Name() string { return c.Model }

func (c *GeminiClient) Generate(ctx context.Context, prompt string) (string, error) {
	// Gemini API는 API 키를 요청 헤더나 쿼리 파라미터로 전달하는 방식을 지원합니다. :contentReference[oaicite:16]{index=16}
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]string{
					{"text": prompt},
				},
			},
		},
	}

	b, _ := json.Marshal(reqBody)
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.Model,
		c.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var decoded struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	json.NewDecoder(resp.Body).Decode(&decoded)

	if len(decoded.Candidates) == 0 {
		return "", errors.New("no candidates found")
	}
	if len(decoded.Candidates[0].Content.Parts) == 0 {
		return "", errors.New("no content parts found")
	}
	return decoded.Candidates[0].Content.Parts[0].Text, nil
}
