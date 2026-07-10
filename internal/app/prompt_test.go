package app

import (
	"os"
	"strings"
	"testing"
)

func TestBuildPromptIncludesWorkflowAndRequirements(t *testing.T) {
	prompt, err := buildPrompt([]string{"NVDA"})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"# gacha auto",
		"Workflow library:",
		"# gacha entry",
		"User request:\nNVDA",
		"Always use current web search",
		"Easy Basic Report",
		"Source Log",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt did not contain %q", expected)
		}
	}
}

func TestBuildPromptRejectsMalformedResearchProfileConfig(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	configDir := dir + "/gacha"
	if err := os.MkdirAll(configDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(configDir+"/config.json", []byte("{invalid"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := buildPrompt([]string{"Should I buy NVDA?"}); err == nil || !strings.Contains(err.Error(), "could not load research profile config") {
		t.Fatalf("expected malformed config error, got %v", err)
	}
}

func TestBuildPromptAutoClassifiesAndRequiresFreshData(t *testing.T) {
	prompt, err := buildPrompt([]string{"NVDA", "지금", "살까?"})
	if err != nil {
		t.Fatal(err)
	}

	for _, expected := range []string{
		"# gacha auto",
		"Classify the user's request",
		"Always use current web search",
		"even if the user does not ask",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt did not contain %q", expected)
		}
	}
}

func TestBuildPromptUsesKoreanForKoreanQuestion(t *testing.T) {
	prompt, err := buildPrompt([]string{"NVDA", "지금", "사도", "될까?"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, "Write the final report in Korean") {
		t.Fatalf("prompt did not request Korean response")
	}
}

func TestBuildPromptLocksReportStructure(t *testing.T) {
	prompt, err := buildPrompt([]string{"What should I buy?"})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"Report structure contract:",
		"The final answer must start with the Easy Basic Report",
		"Write for ordinary users, not investment professionals.",
		"The Basic Report is decision-ready",
		"The Detailed Analysis is the verification layer",
		"Include Detailed Analysis only when",
		"Use simple tables only when they make the answer easier to compare",
		"## Easy Basic Report",
		"### 3. Decision Rules",
		"### 5. Data Check",
		"### 6. More Detail Options",
		"the user knows what they can ask for next",
		"## Detailed Analysis",
		"### G. Source Log",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt did not contain report contract detail %q", expected)
		}
	}
}

func TestBuildPromptIncludesResearchProfile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := saveGachaConfig(gachaConfig{Profile: gachaProfile{
		Markets: profileMulti{
			Selected: []string{profileMarketUSStocksETFs, profileMarketKoreanStocksETFs},
			Default:  profileMarketUSStocksETFs,
		},
		Horizons: profileMulti{
			Selected: []string{profileHorizonSixToTwelve},
			Default:  profileHorizonSixToTwelve,
		},
		Risk:        profileRiskBalanced,
		ReportStyle: profileReportBasicFirst,
		Goals: profileMulti{
			Selected: []string{profileGoalTheme, profileGoalEntry},
			Default:  profileGoalEntry,
		},
		Onboarding: profileOnboarding{Completed: true},
	}}); err != nil {
		t.Fatal(err)
	}
	prompt, err := buildPrompt([]string{"Should I buy NVDA now?"})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"User research profile:",
		"Markets of interest: US stocks / ETFs, Korean stocks / ETFs",
		"Default horizon when unspecified: 6-12 months",
		"Risk preference: Balanced",
		"prefer the user's current question over the saved profile",
		"User request:\nShould I buy NVDA now?",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt missing profile detail %q:\n%s", expected, prompt)
		}
	}
}

func TestBuildPromptOmitsSkippedResearchProfile(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := saveGachaConfig(gachaConfig{Profile: gachaProfile{
		Markets:    profileMulti{Selected: []string{profileMarketUSStocksETFs}, Default: profileMarketUSStocksETFs},
		Onboarding: profileOnboarding{Skipped: true},
	}}); err != nil {
		t.Fatal(err)
	}
	prompt, err := buildPrompt([]string{"Should I buy NVDA now?"})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(prompt, "User research profile:") {
		t.Fatalf("skipped profile should be omitted:\n%s", prompt)
	}
}

func TestBuildPromptUsesKoreanResearchProfileLabels(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	if err := saveGachaConfig(gachaConfig{Profile: gachaProfile{
		Markets: profileMulti{
			Selected: []string{profileMarketKoreanStocksETFs},
			Default:  profileMarketKoreanStocksETFs,
		},
		Horizons: profileMulti{
			Selected: []string{profileHorizonSixToTwelve},
			Default:  profileHorizonSixToTwelve,
		},
		Risk:        profileRiskBalanced,
		ReportStyle: profileReportBasicFirst,
		Goals:       profileMulti{Selected: []string{profileGoalEntry}, Default: profileGoalEntry},
		Onboarding:  profileOnboarding{Completed: true},
	}}); err != nil {
		t.Fatal(err)
	}
	prompt, err := buildPrompt([]string{"삼성전자 지금 사도 될까?"})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"한국 주식/ETF", "6-12개월", "균형형", "매수 진입 계획"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("prompt missing Korean profile label %q:\n%s", expected, prompt)
		}
	}
}

func TestBuildDetailedPromptContinuesFromBasicReport(t *testing.T) {
	prompt, err := buildDetailedPrompt("NVDA 지금 사도 될까?", "## Easy Basic Report\n\nWatch")
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"# gacha detailed analysis",
		"User request:\nNVDA 지금 사도 될까?",
		"Existing basic report:",
		"Write the detailed analysis in Korean",
		`Produce only the "Detailed Analysis" section`,
		"Use these headings in order",
		"Use current web search or current market-data tools again",
	} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("detailed prompt did not contain %q", expected)
		}
	}
}
