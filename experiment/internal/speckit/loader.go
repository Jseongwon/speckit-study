package speckit

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Task 한 건의 태스크 정의
type Task struct {
	Name             string            `yaml:"name"`
	Description      string            `yaml:"description"`
	Inputs           map[string]string `yaml:"inputs"`
	RequiredSections []string          `yaml:"required_sections"`
}

// TaskFile tasks.yaml 최상위 구조
type TaskFile struct {
	Tasks []Task `yaml:"tasks"`
}

// LoadTasks 는 tasks.yaml 파일 경로를 받아 TaskFile 을 반환합니다.
func LoadTasks(path string) (*TaskFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tasks.yaml: %w", err)
	}
	var tf TaskFile
	if err := yaml.Unmarshal(b, &tf); err != nil {
		return nil, fmt.Errorf("parse tasks.yaml: %w", err)
	}
	return &tf, nil
}
