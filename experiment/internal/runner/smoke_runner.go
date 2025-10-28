// internal/runner/smoke_runner.go
package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"speckit-study/internal/llm"
)

type SmokeTaskInput struct {
	ID                 string
	Prompt             string
	CandidateModelTags []string
	IterationsPerTier  int // 보통 3
}

func RunSmokeTest(
	ctx context.Context,
	reg *llm.ModelRegistry,
	in SmokeTaskInput,
) error {

	tsDir := time.Now().Format("20060102_150405")
	baseDir := filepath.Join(".specify", "_runs", in.ID, tsDir)
	os.MkdirAll(baseDir, 0o755)

	for _, tag := range in.CandidateModelTags {
		model, ok := reg.GetModel(tag)
		if !ok {
			fmt.Printf("[SKIP] tag=%s (no model registered)\n", tag)
			continue
		}

		for i := 1; i <= in.IterationsPerTier; i++ {
			out, err := model.Generate(ctx, in.Prompt)
			if err != nil {
				out = fmt.Sprintf("ERROR calling model %s: %v", model.Name(), err)
			}

			filePath := filepath.Join(
				baseDir,
				fmt.Sprintf("%s-try%d-%s.md", tag, i, model.Name()),
			)
			os.WriteFile(filePath, []byte(out), 0o644)

			fmt.Printf("[OK] %s => saved %s\n", tag, filePath)
		}
	}

	return nil
}
