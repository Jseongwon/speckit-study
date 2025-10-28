package runner

import (
	"context"
	"fmt"
	"speckit-study/internal/llm"
)

func main() {
	reg := llm.NewModelRegistry()
	reg.RegisterModel("default", llm.NewOpenAIClient("gpt-4o-mini"))
	reg.RegisterModel("claude", llm.NewAnthropicClient("claude-3-5-sonnet-20240620"))
	reg.RegisterModel("gemini", llm.NewGeminiClient("gemini-2.5-flash"))

	model := reg.DefaultModel()
	if model == nil {
		fmt.Println("⚠️ No default model registered")
		return
	}

	result, err := model.Generate(context.Background(), "Describe SpecKit in one sentence.")
	if err != nil {
		fmt.Println("❌ Error:", err)
		return
	}

	fmt.Println("✅ Model:", model.Name())
	fmt.Println("🧠 Result:", result)
}
