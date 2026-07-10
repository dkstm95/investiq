package app

import (
	"fmt"
	"os"
	"strings"
)

func runPrompt(query string) promptRunResult {
	return runPromptWithProgress(query, nil)
}

func runPromptWithProgress(query string, onOutput func(string)) promptRunResult {
	prompt, err := buildPrompt([]string{query})
	if err != nil {
		return promptRunResult{err: err}
	}
	commandPath, ok := resolveCommand(openCodeCommand)
	if !ok || !hasOpenCodeAuth() {
		return promptRunResult{output: prompt}
	}
	output, err := runOpenCodeWithResolutionProgress(commandPath, prompt, resolveOpenCodeModel(), false, onOutput)
	if err != nil {
		return promptRunResult{output: fallbackPromptOutput(output, prompt), err: err}
	}
	report := strings.TrimSpace(stripANSI(output))
	return promptRunResult{output: report, completed: report != ""}
}

func runDetailedPrompt(query string, basicReport string) promptRunResult {
	return runDetailedPromptWithProgress(query, basicReport, nil)
}

func runDetailedPromptWithProgress(query string, basicReport string, onOutput func(string)) promptRunResult {
	prompt, err := buildDetailedPrompt(query, basicReport)
	if err != nil {
		return promptRunResult{err: err}
	}
	commandPath, ok := resolveCommand(openCodeCommand)
	if !ok || !hasOpenCodeAuth() {
		return promptRunResult{output: prompt}
	}
	output, err := runOpenCodeWithResolutionProgress(commandPath, prompt, resolveOpenCodeModel(), false, onOutput)
	if err != nil {
		return promptRunResult{output: fallbackPromptOutput(output, prompt), err: err}
	}
	report := strings.TrimSpace(stripANSI(output))
	if report == "" {
		return promptRunResult{}
	}
	return promptRunResult{output: strings.TrimSpace(basicReport) + "\n\n" + report, completed: true}
}

func fallbackPromptOutput(diagnostics string, prompt string) string {
	parts := make([]string, 0, 2)
	if diagnostics = strings.TrimSpace(stripANSI(diagnostics)); diagnostics != "" {
		parts = append(parts, diagnostics)
	}
	parts = append(parts, "Copy/paste fallback prompt:\n\n"+strings.TrimSpace(prompt))
	return strings.Join(parts, "\n\n")
}

func welcomeContent(version string, text uiText, width int, _ int) string {
	lang := detectLanguage()
	config, _ := configWithDefaults()
	blocks := []string{brandLine(lang)}
	if profileHasValues(config.Profile) && !config.Profile.Onboarding.Skipped {
		label := "Profile: "
		if lang == languageKorean {
			label = "프로필: "
		}
		blocks = append(blocks, "", mutedStyle.Render(wrapLine(label+profileSummary(config.Profile, lang), max(24, width-4))))
	} else {
		hint := "No research profile set. Type /profile to personalize reports."
		if lang == languageKorean {
			hint = "투자 프로필이 없습니다. /profile에서 리포트를 개인화하세요."
		}
		blocks = append(blocks, "", mutedStyle.Render(wrapLine(hint, max(24, width-4))))
	}
	if onboarding := onboardingContent(text, width, setupReadiness()); onboarding != "" {
		blocks = append(blocks, "", onboarding)
	}
	prompt := "Ask an investment question."
	suggestions := "Discover opportunities · Compare a theme · Plan an entry"
	if lang == languageKorean {
		prompt = "투자 질문을 입력하세요."
		suggestions = "투자 후보 탐색 · 테마 비교 · 매수 진입 계획"
	}
	blocks = append(blocks, "", titleStyle.Render(prompt), "", faintStyle.Render(wrapLine(suggestions, max(24, width-4))), "", faintStyle.Render("v"+version))
	return strings.Join(blocks, "\n")
}

func onboardingContent(text uiText, width int, state setupState) string {
	if state == setupReady {
		return ""
	}
	titleIndex := 0
	bodyIndex := 1
	actionIndex := 2
	if state == setupProviderMissing {
		titleIndex = 3
		bodyIndex = 4
		actionIndex = 5
	}
	lines := []string{
		warningStyle.Render(text.Onboarding[titleIndex]),
		wrapLine(text.Onboarding[bodyIndex], max(20, width-4)),
		mutedStyle.Render(wrapLine(text.Onboarding[actionIndex], max(20, width-4))),
	}
	return calloutStyle.Width(max(24, width-2)).Render(strings.Join(lines, "\n"))
}

type setupState int

const (
	setupReady setupState = iota
	setupRuntimeMissing
	setupProviderMissing
)

func setupReadiness() setupState {
	if _, ok := resolveCommand(openCodeCommand); !ok {
		return setupRuntimeMissing
	}
	if !hasOpenCodeAuth() {
		return setupProviderMissing
	}
	return setupReady
}

func researchingContent(query string, text uiText) string {
	lines := text.Research(query)
	return strings.Join([]string{
		titleStyle.Render(lines[0]),
		lines[1],
		lines[2],
		"",
		titleStyle.Render(lines[3]),
		lines[4],
		lines[5],
		lines[6],
		lines[7],
		lines[8],
		"",
		mutedStyle.Render(lines[9]),
	}, "\n")
}

func helpContent(text uiText) string {
	lines := []string{titleStyle.Render(text.HelpLines[0])}
	for _, line := range text.HelpLines[1:] {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			lines = append(lines, "")
			continue
		}
		command := fields[0]
		description := strings.TrimSpace(strings.TrimPrefix(line, command))
		lines = append(lines, padRight(command, 10)+wrapIndented(description, 56, strings.Repeat(" ", 10)))
	}
	return strings.Join(lines, "\n")
}

func unknownCommandContent(value string, text uiText) string {
	return fmt.Sprintf(text.UnknownCommand, value) + "\n\n" + helpContent(text)
}

func settingsContent(text uiText) string {
	lines := []string{titleStyle.Render(text.SettingsTitle), settingsSummary(text)}
	return strings.Join(lines, "\n")
}

func settingsOverview() string {
	config, err := configWithDefaults()
	if err != nil {
		return err.Error()
	}
	lines := []string{
		fmt.Sprintf("Language: %s", config.Language),
		fmt.Sprintf("Theme:    %s", configuredThemeSummary(config.Theme)),
		fmt.Sprintf("Active:   %s", detectLanguage()),
	}
	if envLang := strings.TrimSpace(os.Getenv("GACHA_LANG")); envLang != "" {
		lines = append(lines, mutedStyle.Render("GACHA_LANG is overriding the language setting."))
	}
	return strings.Join(lines, "\n")
}

func settingsSummary(text uiText) string {
	config, err := configWithDefaults()
	if err != nil {
		return err.Error()
	}
	lines := []string{
		fmt.Sprintf("Config:   %s", gachaConfigPath()),
		fmt.Sprintf("Language: %s", config.Language),
		fmt.Sprintf("Theme:    %s", configuredThemeSummary(config.Theme)),
		fmt.Sprintf("Active:   %s", detectLanguage()),
		"",
		sectionStyle.Render(text.SettingsCommandsTitle),
		"/language auto",
		"/language en",
		"/language ko",
		"/theme system",
		"/theme dark",
		"/theme light",
		"/theme gacha",
	}
	if envLang := strings.TrimSpace(os.Getenv("GACHA_LANG")); envLang != "" {
		lines = append(lines, mutedStyle.Render("GACHA_LANG is currently overriding the language setting."))
	}
	return strings.Join(lines, "\n")
}

func errorContent(err error, output string, text uiText) string {
	parts := []string{titleStyle.Render(text.ErrorTitle), err.Error()}
	if strings.TrimSpace(output) != "" {
		parts = append(parts, "", strings.TrimSpace(output))
	}
	return strings.Join(parts, "\n")
}
