package env

import (
	"os"
	"strings"
)

type Expander struct{}

func NewExpander() *Expander {
	return &Expander{}
}

func (e *Expander) Expand(s string) string {
	var result strings.Builder
	i := 0

	for i < len(s) {
		if s[i] == '$' && i+1 < len(s) {
			var varName strings.Builder
			j := i + 1

			if s[j] == '{' {
				j++
				for j < len(s) && s[j] != '}' {
					varName.WriteByte(s[j])
					j++
				}
				if j < len(s) && s[j] == '}' {
					j++
					val := os.Getenv(varName.String())
					result.WriteString(val)
					i = j
					continue
				}
			} else if (s[j] >= 'a' && s[j] <= 'z') || (s[j] >= 'A' && s[j] <= 'Z') || s[j] == '_' {
				for j < len(s) && ((s[j] >= 'a' && s[j] <= 'z') || (s[j] >= 'A' && s[j] <= 'Z') || (s[j] >= '0' && s[j] <= '9') || s[j] == '_') {
					varName.WriteByte(s[j])
					j++
				}
				val := os.Getenv(varName.String())
				result.WriteString(val)
				i = j
				continue
			}
		}
		result.WriteByte(s[i])
		i++
	}

	return result.String()
}
