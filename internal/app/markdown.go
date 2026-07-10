package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func renderMarkdownReport(markdown string) string {
	markdown = stripANSI(markdown)
	lines := strings.Split(strings.TrimSpace(markdown), "\n")
	rendered := make([]string, 0, len(lines))
	inCodeBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			rendered = append(rendered, mutedStyle.Render(line))
			continue
		}
		if trimmed == "" {
			rendered = append(rendered, "")
			continue
		}
		if strings.HasPrefix(trimmed, "---") && strings.Trim(trimmed, "- ") == "" {
			rendered = append(rendered, mutedStyle.Render(strings.Repeat("─", 24)))
			continue
		}
		if heading, ok := markdownHeading(trimmed); ok {
			rendered = append(rendered, heading)
			continue
		}
		if item, ok := markdownListItem(trimmed); ok {
			rendered = append(rendered, bulletStyle.Render("•")+" "+inlineMarkdown(item))
			continue
		}
		rendered = append(rendered, inlineMarkdown(line))
	}
	return strings.TrimSpace(strings.Join(rendered, "\n"))
}

func markdownHeading(line string) (string, bool) {
	level := 0
	for level < len(line) && line[level] == '#' {
		level++
	}
	if level == 0 || level > 6 || level >= len(line) || line[level] != ' ' {
		return "", false
	}
	text := inlineMarkdown(strings.TrimSpace(line[level:]))
	if level <= 2 {
		return sectionStyle.Render(text), true
	}
	return titleStyle.Render(text), true
}

func markdownListItem(line string) (string, bool) {
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return strings.TrimSpace(line[2:]), true
	}
	for i, r := range line {
		if r == '.' && i > 0 && i+1 < len(line) && line[i+1] == ' ' {
			for _, digit := range line[:i] {
				if digit < '0' || digit > '9' {
					return "", false
				}
			}
			return strings.TrimSpace(line[i+2:]), true
		}
		if r < '0' || r > '9' {
			return "", false
		}
	}
	return "", false
}

func inlineMarkdown(value string) string {
	value = strings.ReplaceAll(value, "**", "")
	value = strings.ReplaceAll(value, "__", "")
	value = strings.ReplaceAll(value, "`", "")
	value = strings.ReplaceAll(value, "\\[", "[")
	value = strings.ReplaceAll(value, "\\]", "]")
	if lipgloss.Width(value) == 0 {
		return ""
	}
	return value
}
