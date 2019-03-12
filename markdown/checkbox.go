package markdown

import (
	"regexp"
)

var (
	checkboxRegexp     = regexp.MustCompile(`(?i:\[[ x]\])`)
	checkboxLineRegexp = regexp.MustCompile(`(?mi:^\s*\*\s?\[[ x]\](.*)$)`)
)

func toggleSingleCheckbox(t string) string {
	i := 0
	return checkboxRegexp.ReplaceAllStringFunc(t, func(s string) string {
		i++
		if i == 1 {
			if s == "[ ]" {
				return "[X]"
			}
			return "[ ]"
		}
		return s
	})
}

func ToggleCheckbox(t string, n int) string {
	i := 0
	return checkboxLineRegexp.ReplaceAllStringFunc(t, func(s string) string {
		i++
		if i == n {
			return toggleSingleCheckbox(s)
		}
		return s
	})
}
