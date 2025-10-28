// internal/llm/copilot_client.go
package llm

import (
	"context"
	"errors"
)

type CopilotClient struct{}

func NewCopilotClient() *CopilotClient {
	return &CopilotClient{}
}

func (c *CopilotClient) Name() string { return "github-copilot" }

func (c *CopilotClient) Generate(ctx context.Context, prompt string) (string, error) {
	// GitHub Copilot Chat은 일반 개발자용 공개 LLM API 엔드포인트가 현재 제공되지 않습니다.
	// 관리/메트릭용 REST API만 공개돼 있다고 문서와 커뮤니티 Q&A에서 명시되어 있습니다. :contentReference[oaicite:18]{index=18}
	return "", errors.New("Copilot programmatic generation not supported")
}
