package llm

import (
	"context"
	"fmt"
	"sync"
)

// LLMClient 인터페이스는 모든 언어 모델 클라이언트가 구현해야 하는 최소 계약입니다.
type LLMClient interface {
	Name() string
	Generate(ctx context.Context, prompt string) (string, error)
}

// ModelRegistry는 다양한 LLMClient를 태그로 등록/조회할 수 있도록 관리합니다.
// 예: "gpt", "claude", "gemini" 등의 태그를 사용.
type ModelRegistry struct {
	tagToClient map[string]LLMClient
	mu          sync.RWMutex
}

// NewModelRegistry는 새로운 레지스트리를 초기화합니다.
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		tagToClient: make(map[string]LLMClient),
	}
}

// RegisterModel은 모델을 태그와 함께 등록합니다.
// 예: reg.RegisterModel("gpt", NewOpenAIClient())
func (r *ModelRegistry) RegisterModel(tag string, client LLMClient) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tagToClient[tag] = client
}

// GetModel은 태그로 모델을 조회합니다.
// 반환값: (LLMClient, 존재 여부)
func (r *ModelRegistry) GetModel(tag string) (LLMClient, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	c, ok := r.tagToClient[tag]
	return c, ok
}

// DefaultModel은 기본 모델을 반환합니다.
// "default" 태그가 없으면 nil 반환.
func (r *ModelRegistry) DefaultModel() LLMClient {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if c, ok := r.tagToClient["default"]; ok {
		return c
	}
	return nil
}

// MustGetModel은 필수 모델을 조회합니다.
// 없을 경우 panic을 발생시킵니다 (테스트/런타임 디버깅용).
func (r *ModelRegistry) MustGetModel(tag string) LLMClient {
	c, ok := r.GetModel(tag)
	if !ok {
		panic(fmt.Sprintf("LLM model not found for tag: %s", tag))
	}
	return c
}

// ListModels는 현재 등록된 모든 모델 태그를 반환합니다.
func (r *ModelRegistry) ListModels() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	keys := make([]string, 0, len(r.tagToClient))
	for k := range r.tagToClient {
		keys = append(keys, k)
	}
	return keys
}

// Example: 레지스트리 초기화 예시
func ExampleRegistry() {
	reg := NewModelRegistry()
	reg.RegisterModel("default", NewOpenAIClient("gpt-4o-mini"))
	reg.RegisterModel("gpt", NewOpenAIClient("gpt-4o-mini"))
	reg.RegisterModel("claude", NewAnthropicClient("claude-3-5-sonnet-20240620"))
	reg.RegisterModel("gemini", NewGeminiClient("gemini-2.5-flash"))

	// 모델 조회
	model := reg.MustGetModel("default")
	result, _ := model.Generate(context.Background(), "Write a haiku about Go.")
	fmt.Println(result)
}
