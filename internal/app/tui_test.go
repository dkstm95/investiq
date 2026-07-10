package app

import (
	"os"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func keyMsg(value string) tea.KeyMsg {
	switch value {
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEsc}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func skipProfileOnboardingForTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	if err := saveGachaConfig(gachaConfig{Profile: gachaProfile{Onboarding: profileOnboarding{Skipped: true}}}); err != nil {
		t.Fatal(err)
	}
}

func TestTUILanguageCommandIsSettingsOnly(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	model := newTUIModel("0.1.27")
	next, cmd := model.handleSubmit("/language")
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	updated := next.(tuiModel)
	if updated.choice != nil {
		t.Fatalf("language should not open as a top-level selector: %#v", updated.choice)
	}
	got := stripANSI(updated.view.View())
	if !strings.Contains(got, "Unknown command: /language") {
		t.Fatalf("language should be a settings-only command:\n%s", got)
	}
}

func TestTUIUnknownSlashCommandDoesNotRunPrompt(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	next, cmd := model.handleSubmit("/setting")
	if cmd != nil {
		t.Fatal("unknown slash command should not run a prompt command")
	}
	updated := next.(tuiModel)
	got := stripANSI(updated.view.View())
	for _, expected := range []string{"Unknown command: /setting", "Command palette"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("unknown command view missing %q:\n%s", expected, got)
		}
	}
}

func TestTUIShowsPromptOutputWhileBusy(t *testing.T) {
	model := newTUIModel("0.1.27")
	model.busy = true
	model.query = "NVDA"
	model.view.SetContent(researchingContent(model.query, model.text))

	next, cmd := model.Update(promptOutputMsg{query: "NVDA", chunk: "\x1b[32mpartial report\rnext line"})
	if cmd != nil {
		t.Fatal("unexpected command without stream")
	}
	updated := next.(tuiModel)
	got := stripANSI(updated.view.View())
	if !strings.Contains(got, "partial report") || !strings.Contains(got, "next line") {
		t.Fatalf("streamed output not visible:\n%s", got)
	}
	if updated.mode != updated.text.Report {
		t.Fatalf("expected report mode while output is streaming, got %q", updated.mode)
	}
}

func TestTUIBlocksNewResearchWhileBusy(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	model.busy = true
	model.query = "first question"
	model.input.SetValue("second question")

	next, cmd := model.Update(keyMsg("enter"))
	if cmd != nil {
		t.Fatal("busy TUI should not start another research command")
	}
	updated := next.(tuiModel)
	if updated.query != "first question" || updated.input.Value() != "second question" {
		t.Fatalf("busy submission changed active work: query=%q input=%q", updated.query, updated.input.Value())
	}
}

func TestTUIIgnoresStaleRunResult(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	model.busy = true
	model.query = "current question"
	model.report = "current progress"

	next, cmd := model.Update(runResultMsg{query: "old question", output: "stale report", completed: true})
	if cmd != nil {
		t.Fatal("stale result should not schedule a command")
	}
	updated := next.(tuiModel)
	if !updated.busy || updated.query != "current question" || updated.report != "current progress" {
		t.Fatalf("stale result changed current state: %#v", updated)
	}
}

func TestTUIStreamingHoldsSplitOSCSequence(t *testing.T) {
	model := newTUIModel("0.1.27")
	model.busy = true
	model.query = "NVDA"

	next, _ := model.Update(promptOutputMsg{query: "NVDA", chunk: "before\x1b]52;c;payload"})
	updated := next.(tuiModel)
	if got := stripANSI(updated.view.View()); strings.Contains(got, "payload") {
		t.Fatalf("incomplete OSC payload leaked into TUI: %q", got)
	}
	next, _ = updated.Update(promptOutputMsg{query: "NVDA", chunk: "\x07after"})
	updated = next.(tuiModel)
	if got := stripANSI(updated.view.View()); !strings.Contains(got, "beforeafter") || strings.Contains(got, "payload") {
		t.Fatalf("unexpected sanitized stream: %q", got)
	}
}

func TestOnboardingContentReflectsSetupState(t *testing.T) {
	text := englishText()
	if got := onboardingContent(text, 80, setupReady); got != "" {
		t.Fatalf("expected no onboarding for ready state, got %q", got)
	}
	if got := onboardingContent(text, 80, setupRuntimeMissing); !strings.Contains(stripANSI(got), "OpenCode is not installed yet") {
		t.Fatalf("unexpected runtime onboarding: %q", got)
	}
	if got := onboardingContent(text, 80, setupProviderMissing); !strings.Contains(stripANSI(got), "no AI provider") {
		t.Fatalf("unexpected provider onboarding: %q", got)
	}
}

func TestTUIMalformedConfigShowsRecoverableError(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	configDir := dir + "/gacha"
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configDir+"/config.json", []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}

	model := newTUIModel("0.1.27")
	if model.profile != nil {
		t.Fatal("malformed config should not trap the user in profile onboarding")
	}
	got := stripANSI(model.view.View())
	if !strings.Contains(got, "could not load config") {
		t.Fatalf("config recovery error was not visible:\n%s", got)
	}
	if _, cmd := model.handleSubmit("/quit"); cmd == nil {
		t.Fatal("typed quit should remain available after a config error")
	}
}

func TestFallbackPromptOutputSanitizesDiagnosticsAndKeepsPrompt(t *testing.T) {
	got := fallbackPromptOutput("\x1b]52;c;payload\x07runtime failed", "full research prompt")
	if strings.Contains(got, "payload") || !strings.Contains(got, "runtime failed") || !strings.Contains(got, "full research prompt") {
		t.Fatalf("unexpected fallback output: %q", got)
	}
}

func TestWelcomeContentIsPromptFirst(t *testing.T) {
	skipProfileOnboardingForTest(t)
	got := stripANSI(welcomeContent("0.1.27", englishText(), 80, 16))
	for _, expected := range []string{
		"GACHA",
		"Better odds through research.",
		"Ask an investment question.",
		"No research profile set. Type /profile",
		"Discover opportunities",
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("welcome content missing %q:\n%s", expected, got)
		}
	}
	for _, unwanted := range []string{
		"Ask -> Fresh data",
		"[Fresh data required]",
		"[No trading]",
		"Decision desk",
	} {
		if strings.Contains(got, unwanted) {
			t.Fatalf("welcome content kept decorative element %q:\n%s", unwanted, got)
		}
	}
}

func TestWelcomeContentFitsCommonTerminalSizes(t *testing.T) {
	skipProfileOnboardingForTest(t)
	for _, tc := range []struct {
		name   string
		width  int
		height int
		text   uiText
	}{
		{name: "quarter english", width: 80, height: 16, text: englishText()},
		{name: "half english", width: 170, height: 24, text: englishText()},
		{name: "quarter korean", width: 80, height: 16, text: koreanText()},
		{name: "half korean", width: 170, height: 24, text: koreanText()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := stripANSI(welcomeContent("0.1.27", tc.text, tc.width, tc.height))
			for _, line := range strings.Split(got, "\n") {
				if lipgloss.Width(line) > tc.width {
					t.Fatalf("line width %d exceeds %d: %q\n%s", lipgloss.Width(line), tc.width, line, got)
				}
			}
		})
	}
}

func TestTUIViewFitsCommonTerminalSizes(t *testing.T) {
	for _, tc := range []struct {
		name   string
		width  int
		height int
	}{
		{name: "quarter", width: 80, height: 24},
		{name: "half", width: 120, height: 30},
		{name: "full", width: 180, height: 36},
	} {
		t.Run(tc.name, func(t *testing.T) {
			skipProfileOnboardingForTest(t)
			model := newTUIModel("0.1.27")
			model.width = tc.width
			model.height = tc.height
			model.view.Width = max(30, tc.width-8)
			model.view.Height = max(6, tc.height-8)
			got := stripANSI(model.View())
			for _, line := range strings.Split(got, "\n") {
				if lipgloss.Width(line) > tc.width {
					t.Fatalf("line width %d exceeds %d: %q\n%s", lipgloss.Width(line), tc.width, line, got)
				}
			}
			if strings.Contains(got, "Checks fresh data") {
				t.Fatalf("status bar repeated safety copy:\n%s", got)
			}
			if tc.name == "full" {
				for _, expected := range []string{"GACHA", "Better odds through research.", "Ask an investment question.", "Ready"} {
					if !strings.Contains(got, expected) {
						t.Fatalf("full layout missing %q:\n%s", expected, got)
					}
				}
			}
		})
	}
}

func TestTUICommandViewsFitQuarterTerminal(t *testing.T) {
	for _, command := range []string{"/theme", "/help"} {
		t.Run(command, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			model := newTUIModel("0.1.27")
			next, cmd := model.handleSubmit(command)
			if cmd != nil {
				t.Fatal("unexpected command")
			}
			updated := next.(tuiModel)
			updated.width = 80
			updated.height = 24
			got := stripANSI(updated.View())
			for _, line := range strings.Split(got, "\n") {
				if lipgloss.Width(line) > 80 {
					t.Fatalf("line width %d exceeds 80: %q\n%s", lipgloss.Width(line), line, got)
				}
			}
			for _, fragment := range []string{
				"│  pro                                                                     │",
				"│  /languag                                                                │",
			} {
				if strings.Contains(got, fragment) {
					t.Fatalf("command view contains wrapped fragment %q:\n%s", fragment, got)
				}
			}
		})
	}
}

func TestTUIChoiceViewsRerenderAfterNarrowResize(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	model := newTUIModel("0.1.27")
	next, cmd := model.handleSubmit("/theme")
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	resized, cmd := next.(tuiModel).Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	got := stripANSI(resized.(tuiModel).View())
	for _, line := range strings.Split(got, "\n") {
		if lipgloss.Width(line) > 60 {
			t.Fatalf("line width %d exceeds 60: %q\n%s", lipgloss.Width(line), line, got)
		}
	}
	for _, fragment := range []string{
		"│  backgroun",
		"│  lig",
		"│  Gach",
		"──                                                    │",
	} {
		if strings.Contains(got, fragment) {
			t.Fatalf("narrow theme view contains clipped fragment %q:\n%s", fragment, got)
		}
	}
}

func TestTUIFullLayoutKeepsLocalizedPromptInRightPanel(t *testing.T) {
	for _, tc := range []struct {
		name       string
		text       uiText
		promptText string
	}{
		{name: "english", text: englishText(), promptText: "Ask:"},
		{name: "korean", text: koreanText(), promptText: "예:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			skipProfileOnboardingForTest(t)
			model := newTUIModel("0.1.27")
			model.text = tc.text
			model.input.Placeholder = tc.text.InputPlaceholder
			model.width = 180
			model.height = 36
			got := stripANSI(model.View())
			if !strings.Contains(got, tc.promptText) {
				t.Fatalf("missing localized prompt %q:\n%s", tc.promptText, got)
			}
		})
	}
}

func TestTUIStatusBarRendersBelowWorkspace(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	model.width = 180
	model.height = 36
	got := stripANSI(model.View())
	promptLine := -1
	statusLine := -1
	panelBottomLine := -1
	for i, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "Ask:") {
			promptLine = i
		}
		if strings.Contains(line, "Ready") && strings.Contains(line, "Mode") {
			statusLine = i
		}
		if strings.Contains(line, "╰") && statusLine < 0 {
			panelBottomLine = i
		}
	}
	if promptLine < 0 {
		t.Fatalf("missing prompt input:\n%s", got)
	}
	if statusLine <= promptLine {
		t.Fatalf("status bar did not render below workspace; prompt line %d status line %d:\n%s", promptLine, statusLine, got)
	}
	if panelBottomLine < 0 {
		t.Fatalf("missing panel bottom border:\n%s", got)
	}
}

func TestTUIOnboardingStartsWhenProfileMissing(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	model := newTUIModel("0.1.27")
	model.width = 180
	model.height = 36

	got := stripANSI(model.View())
	for _, expected := range []string{"Better odds through research.", "Primary markets", "space toggle"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("onboarding missing %q:\n%s", expected, got)
		}
	}
	researchLine := -1
	profileLine := -1
	lines := strings.Split(got, "\n")
	for i, line := range lines {
		if strings.Contains(line, "disciplined research.") {
			researchLine = i
		}
		if strings.Contains(line, "Let's set your research profile.") {
			profileLine = i
		}
	}
	if researchLine < 0 || profileLine != researchLine+2 || strings.Trim(strings.Trim(lines[researchLine+1], "│"), " ") != "" {
		t.Fatalf("onboarding intro should keep separate paragraphs:\n%s", got)
	}
	if strings.Contains(got, "enter run") {
		t.Fatalf("onboarding status should not show the global prompt footer:\n%s", got)
	}
}

func TestTUIOnboardingCanSaveProfile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	model := newTUIModel("0.1.27")
	for _, key := range []string{" ", "enter", " ", "enter", "enter", "enter", " ", "enter"} {
		next, cmd := model.Update(keyMsg(key))
		if cmd != nil {
			t.Fatal("unexpected command")
		}
		model = next.(tuiModel)
	}
	config, err := loadGachaConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !config.Profile.Onboarding.Completed || config.Profile.Onboarding.Skipped {
		t.Fatalf("unexpected onboarding state: %#v", config.Profile.Onboarding)
	}
	if config.Profile.Markets.Default != profileMarketUSStocksETFs || config.Profile.Horizons.Default != profileHorizonOneToThreeMonths {
		t.Fatalf("unexpected saved profile: %#v", config.Profile)
	}
	if config.Profile.Risk != profileRiskBalanced || config.Profile.ReportStyle != profileReportBasicFirst {
		t.Fatalf("unexpected saved single-select profile: %#v", config.Profile)
	}
}

func TestTUIOnboardingSkipSavesSkippedState(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	model := newTUIModel("0.1.27")
	for range profileOnboardingSteps {
		next, cmd := model.Update(keyMsg("s"))
		if cmd != nil {
			t.Fatal("unexpected command")
		}
		model = next.(tuiModel)
	}
	config, err := loadGachaConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.Profile.Onboarding.Completed || !config.Profile.Onboarding.Skipped {
		t.Fatalf("expected skipped onboarding state, got %#v", config.Profile.Onboarding)
	}
	model = newTUIModel("0.1.27")
	if model.profile != nil {
		t.Fatalf("skipped onboarding should not reopen, got %#v", model.profile)
	}
}

func TestTUIOnboardingRejectsEmptyMultiSelect(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	model := newTUIModel("0.1.27")
	next, cmd := model.Update(keyMsg("enter"))
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	updated := next.(tuiModel)
	if updated.profile == nil || updated.profile.Category != profileCategoryMarkets {
		t.Fatalf("expected onboarding to stay on markets, got %#v", updated.profile)
	}
	got := stripANSI(updated.View())
	if !strings.Contains(got, "Select at least one option") {
		t.Fatalf("missing inline validation:\n%s", got)
	}
	config, err := loadGachaConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !profileIsZero(config.Profile) {
		t.Fatalf("invalid category should not save profile: %#v", config.Profile)
	}
}

func TestTUIProfileCommandOpensEditor(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	next, cmd := model.handleSubmit("/profile")
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	updated := next.(tuiModel)
	if updated.profile == nil || updated.profile.Mode != profileFlowMenu {
		t.Fatalf("expected profile menu, got %#v", updated.profile)
	}
	got := stripANSI(updated.view.View())
	for _, expected := range []string{"Research Profile", "Edit markets", "Reset profile"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("profile editor missing %q:\n%s", expected, got)
		}
	}
}

func TestTUIProfileEditorPersistsEditedCategory(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	next, cmd := model.handleSubmit("/profile")
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	model = next.(tuiModel)
	for _, key := range []string{"enter", " ", "enter"} {
		next, cmd = model.Update(keyMsg(key))
		if cmd != nil {
			t.Fatal("unexpected command")
		}
		model = next.(tuiModel)
	}
	config, err := loadGachaConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !config.Profile.Onboarding.Completed || config.Profile.Onboarding.Skipped {
		t.Fatalf("edited profile should become active, got %#v", config.Profile.Onboarding)
	}
	if config.Profile.Markets.Default != profileMarketUSStocksETFs {
		t.Fatalf("expected edited market to persist, got %#v", config.Profile.Markets)
	}
}

func TestTUIProfileEditorResetPreservesSystemSettings(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := saveGachaConfig(gachaConfig{
		Language: languageSettingKorean,
		Theme:    themeSettingGacha,
		Profile: gachaProfile{
			Markets:    profileMulti{Selected: []string{profileMarketUSStocksETFs}, Default: profileMarketUSStocksETFs},
			Risk:       profileRiskBalanced,
			Onboarding: profileOnboarding{Completed: true},
		},
	}); err != nil {
		t.Fatal(err)
	}
	model := newTUIModel("0.1.27")
	next, cmd := model.handleSubmit("/profile")
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	model = next.(tuiModel)
	for i := 0; i < len(profileMenuLabels(languageEnglish))-1; i++ {
		next, cmd = model.Update(keyMsg("down"))
		if cmd != nil {
			t.Fatal("unexpected command")
		}
		model = next.(tuiModel)
	}
	next, cmd = model.Update(keyMsg("enter"))
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	config, err := loadGachaConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.Language != languageSettingKorean || config.Theme != themeSettingGacha {
		t.Fatalf("reset changed system settings: %#v", config)
	}
	if !profileIsZero(config.Profile) {
		t.Fatalf("reset should clear profile, got %#v", config.Profile)
	}
}

func TestTUIProfileKeysDoNotExit(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	model := newTUIModel("0.1.27")
	if model.profile == nil || model.profile.Mode != profileFlowOnboarding {
		t.Fatalf("expected onboarding profile flow, got %#v", model.profile)
	}
	for _, key := range []string{"esc", "ctrl+c"} {
		next, cmd := model.Update(keyMsg(key))
		if cmd != nil {
			t.Fatalf("%s should not quit profile onboarding", key)
		}
		updated := next.(tuiModel)
		if updated.profile == nil {
			t.Fatalf("%s should keep profile onboarding open", key)
		}
	}
}

func TestTUIProfileEscBacksOutOfEditor(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	next, cmd := model.handleSubmit("/profile")
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	model = next.(tuiModel)
	next, cmd = model.Update(keyMsg("enter"))
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	model = next.(tuiModel)
	if model.profile == nil || model.profile.Mode != profileFlowEdit {
		t.Fatalf("expected profile edit flow, got %#v", model.profile)
	}
	next, cmd = model.Update(keyMsg("esc"))
	if cmd != nil {
		t.Fatal("esc should not quit from profile edit")
	}
	model = next.(tuiModel)
	if model.profile == nil || model.profile.Mode != profileFlowMenu {
		t.Fatalf("expected esc to return to profile menu, got %#v", model.profile)
	}
	next, cmd = model.Update(keyMsg("esc"))
	if cmd != nil {
		t.Fatal("esc should close profile menu without quitting")
	}
	model = next.(tuiModel)
	if model.profile != nil {
		t.Fatalf("expected esc to close profile menu, got %#v", model.profile)
	}
}

func TestReportActionsExposeNextChoices(t *testing.T) {
	got := stripANSI(renderReportActions(englishText()))
	normalized := strings.Join(strings.Fields(got), " ")
	for _, expected := range []string{"Next", "Use ↑/↓ and enter", "Detailed analysis", "Save report", "Skip saving", "Shortcut: d", "ask a new question"} {
		if !strings.Contains(normalized, expected) {
			t.Fatalf("report actions missing %q:\n%s", expected, got)
		}
	}
}

func TestCompletedReportDoesNotAppendActionChoices(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	model.query = "Should I buy NVDA now?"
	next, cmd := model.Update(runResultMsg{
		query:     "Should I buy NVDA now?",
		output:    "## Easy Basic Report\n\n### 1. Bottom Line\nRead this first.",
		completed: true,
	})
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	updated := next.(tuiModel)
	got := stripANSI(updated.view.View())
	if updated.choice != nil {
		t.Fatalf("report should not enter choice mode immediately: %#v", updated.choice)
	}
	for _, expected := range []string{"Easy Basic Report", "Read this first.", "Next:"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("report view missing %q:\n%s", expected, got)
		}
	}
	for _, unexpected := range []string{"Use ↑/↓ and enter", "Save report", "Skip saving"} {
		if strings.Contains(got, unexpected) {
			t.Fatalf("report view unexpectedly included action choice %q:\n%s", unexpected, got)
		}
	}
}

func TestReportActionsOpenOnEnterAndEscReturnsToReport(t *testing.T) {
	skipProfileOnboardingForTest(t)
	model := newTUIModel("0.1.27")
	model.report = "## Easy Basic Report\n\n### 1. Bottom Line\nRead this first."
	model.save = &pendingSave{query: "Should I buy NVDA now?", report: model.report}
	model.view.SetContent(reportContentWithPrompt(model.report, model.text))

	next, cmd := model.Update(keyMsg("enter"))
	if cmd != nil {
		t.Fatal("unexpected command")
	}
	actions := next.(tuiModel)
	if actions.choice == nil || actions.choice.Kind != choiceReport {
		t.Fatalf("expected report action choice, got %#v", actions.choice)
	}
	actionView := stripANSI(actions.view.View())
	if strings.Contains(actionView, "Read this first.") {
		t.Fatalf("action screen should not include report body:\n%s", actionView)
	}
	if !strings.Contains(actionView, "Save report") {
		t.Fatalf("action screen missing save choice:\n%s", actionView)
	}

	next, cmd = actions.Update(keyMsg("esc"))
	if cmd != nil {
		t.Fatal("esc in report actions should return to the report without quitting")
	}
	report := next.(tuiModel)
	if report.choice != nil {
		t.Fatalf("expected choice mode to close, got %#v", report.choice)
	}
	reportView := stripANSI(report.view.View())
	if !strings.Contains(reportView, "Read this first.") {
		t.Fatalf("report view missing original body after esc:\n%s", reportView)
	}
}

func TestTUIOnlyTypedQuitCommandsExit(t *testing.T) {
	skipProfileOnboardingForTest(t)
	for _, key := range []string{"esc", "ctrl+c"} {
		t.Run(key, func(t *testing.T) {
			model := newTUIModel("0.1.27")
			next, cmd := model.Update(keyMsg(key))
			if cmd != nil {
				t.Fatalf("%s should not quit the TUI", key)
			}
			if _, ok := next.(tuiModel); !ok {
				t.Fatalf("%s returned unexpected model type %T", key, next)
			}
		})
	}

	model := newTUIModel("0.1.27")
	next, cmd := model.handleSubmit("exit")
	if cmd == nil {
		t.Fatal("typed exit command should quit the TUI")
	}
	if _, ok := next.(tuiModel); !ok {
		t.Fatalf("exit returned unexpected model type %T", next)
	}
}

func TestReportContextRecognizesKoreanSections(t *testing.T) {
	got := reportContextFromMarkdown("## 쉬운 기본 리포트\n\n### 1. 결론\n대기.\n\n### 3. 행동 기준\n조건.\n\n### 4. 리스크\n높은 밸류에이션.\n\n### 5. 데이터 확인\n출처 포함.", "fallback")
	for _, expected := range []string{"결론", "행동 기준", "리스크", "데이터"} {
		found := false
		for _, item := range got {
			if item == expected {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing Korean report context %q in %#v", expected, got)
		}
	}
}

func TestReportMarkdownRendersForDisplay(t *testing.T) {
	got := stripANSI(renderMarkdownReport("## Easy Basic Report\n\n### 1. Bottom Line\n- **Wait** for pullback\n\n---\n\n`NVDA`"))
	for _, unexpected := range []string{"##", "###", "**", "`", "---"} {
		if strings.Contains(got, unexpected) {
			t.Fatalf("rendered report still contains markdown marker %q:\n%s", unexpected, got)
		}
	}
	for _, expected := range []string{"Easy Basic Report", "1. Bottom Line", "• Wait for pullback", "NVDA"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("rendered report missing %q:\n%s", expected, got)
		}
	}
}

func TestReportChoiceStaysSeparateFromRenderedReport(t *testing.T) {
	model := newTUIModel("0.1.27")
	model.mode = model.text.Report
	model.report = "## Easy Basic Report\n\n### 1. Bottom Line\nWait."
	model = model.showReportChoice()
	model.moveChoice(1)

	got := stripANSI(model.view.View())
	for _, unexpected := range []string{"Easy Basic Report", "Wait.", "##", "###"} {
		if strings.Contains(got, unexpected) {
			t.Fatalf("report choice unexpectedly included report body %q:\n%s", unexpected, got)
		}
	}
	if !strings.Contains(got, "Save report") {
		t.Fatalf("report choice missing actions:\n%s", got)
	}
}

func TestTUIHelpExposesOnlyUserFacingCommands(t *testing.T) {
	got := stripANSI(helpContent(englishText()))
	for _, expected := range []string{"/home", "/help", "/profile", "/settings", "/theme", "/quit"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("help content missing %q:\n%s", expected, got)
		}
	}
	for _, expected := range []string{"return to the prompt", "language and theme settings"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("help content missing updated copy %q:\n%s", expected, got)
		}
	}
	for _, hidden := range []string{"/doctor", "/setup", "/update", "/model", "/language"} {
		if strings.Contains(got, hidden) {
			t.Fatalf("help content exposed operational command %q:\n%s", hidden, got)
		}
	}
	if strings.Contains(got, "dashboard") {
		t.Fatalf("help content kept outdated dashboard copy:\n%s", got)
	}
}

func TestFooterKeepsPrimaryCommandsVisible(t *testing.T) {
	got := stripANSI(renderFooter(120, englishText()))
	for _, expected := range []string{"/profile", "/settings", "/quit"} {
		if !strings.Contains(got, expected) {
			t.Fatalf("footer missing %q:\n%s", expected, got)
		}
	}
	for _, hidden := range []string{"/doctor", "/setup", "/update"} {
		if strings.Contains(got, hidden) {
			t.Fatalf("footer exposed operational command %q:\n%s", hidden, got)
		}
	}
	if strings.Contains(strings.ToLower(got), "esc") {
		t.Fatalf("footer should not advertise esc as an exit command:\n%s", got)
	}
}
