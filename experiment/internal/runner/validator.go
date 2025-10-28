package runner

import (
	"regexp"
	"strings"
)

// ValidateRequiredSections 는 마크다운 문서 내에 필수 섹션(헤딩)이 존재하는지 검사합니다.
// 섹션 이름은 대소문자 무시, "## <Name>" 형태를 기본으로 탐지합니다.
func ValidateRequiredSections(markdown string, required []string) (missing []string) {
	if len(required) == 0 {
		return nil
	}
	md := "\n" + markdown + "\n"
	for _, sec := range required {
		// ^##\s*Goal(\s|$)  형태로 탐지 (대소문자 구분 없음)
		pat := `(?m)^\s{0,3}#{2,6}\s*` + regexp.QuoteMeta(sec) + `(\s|$)`
		re := regexp.MustCompile(`(?i)` + pat)
		if !re.MatchString(md) {
			missing = append(missing, sec)
		}
	}
	return
}

// HasAllRequiredSections : 전부 만족하는지 boolean 반환 헬퍼
func HasAllRequiredSections(markdown string, required []string) bool {
	return len(ValidateRequiredSections(markdown, required)) == 0
}

// NormalizeNewlines : 플랫폼간 개행 통일(테스트 안정화용)
func NormalizeNewlines(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	return s
}
