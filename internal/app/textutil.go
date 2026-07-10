package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

func settingValue(value string) string {
	fields := strings.Fields(value)
	if len(fields) < 2 {
		return ""
	}
	return strings.TrimSpace(fields[1])
}

func configuredThemeSummary(theme string) string {
	normalized, ok := normalizeThemeSetting(theme)
	if !ok {
		return themeSettingSystem
	}
	return normalized
}

func wrapLine(value string, width int) string {
	value = strings.TrimSpace(value)
	if width <= 0 || lipgloss.Width(value) <= width {
		return value
	}
	words := strings.Fields(value)
	if len(words) == 0 {
		return value
	}
	var lines []string
	current := words[0]
	for _, word := range words[1:] {
		next := current + " " + word
		if lipgloss.Width(next) > width {
			lines = append(lines, current)
			current = word
			continue
		}
		current = next
	}
	lines = append(lines, current)
	return strings.Join(lines, "\n")
}

func wrapParagraphs(value string, width int) string {
	paragraphs := strings.Split(strings.TrimSpace(value), "\n\n")
	wrapped := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		wrapped = append(wrapped, wrapLine(paragraph, width))
	}
	return strings.Join(wrapped, "\n\n")
}

func wrapIndented(value string, width int, indent string) string {
	wrapped := wrapLine(value, width)
	lines := strings.Split(wrapped, "\n")
	for i := 1; i < len(lines); i++ {
		lines[i] = indent + lines[i]
	}
	return strings.Join(lines, "\n")
}

func padRight(value string, width int) string {
	padding := width - lipgloss.Width(value)
	if padding <= 0 {
		return value
	}
	return value + strings.Repeat(" ", padding)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func stripANSI(value string) string {
	return ansi.Strip(value)
}
