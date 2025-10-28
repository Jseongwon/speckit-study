package speckit

import (
	"fmt"
	"sort"
	"strings"
)

// BuildPrompt 는 specify.md + plan.md + inputs 를 하나의 프롬프트로 합칩니다.
func BuildPrompt(specify string, plan string, inputs map[string]string) string {
	var sb strings.Builder
	sb.WriteString("# System\nYou are a senior software engineer helping to materialize a SpecKit plan.\n\n")

	// Inputs (정렬 출력)
	if len(inputs) > 0 {
		keys := make([]string, 0, len(inputs))
		for k := range inputs {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		sb.WriteString("## Inputs\n")
		for _, k := range keys {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", k, inputs[k]))
		}
		sb.WriteString("\n")
	}

	// specify
	if strings.TrimSpace(specify) != "" {
		sb.WriteString("## Specification\n")
		sb.WriteString(specify)
		sb.WriteString("\n\n")
	}

	// plan
	if strings.TrimSpace(plan) != "" {
		sb.WriteString("## Plan\n")
		sb.WriteString(plan)
		sb.WriteString("\n\n")
	}

	// Output format 힌트
	sb.WriteString("## Output Requirements\n")
	sb.WriteString("- Keep the answer concise and directly usable by developers.\n")
	sb.WriteString("- Use markdown when appropriate.\n")
	return sb.String()
}
