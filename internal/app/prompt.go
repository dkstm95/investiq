package app

import (
	"bytes"
	"fmt"
	"io/fs"
	"strings"

	"github.com/dkstm95/gacha/internal/agent"
)

func printPrompt(query []string) error {
	prompt, err := buildPrompt(query)
	if err != nil {
		return err
	}
	fmt.Println(prompt)
	return nil
}

func runQuery(args []string) error {
	return New("").runQuery(args)
}

func (a *App) runQuery(args []string) error {
	dryRun := false
	query := make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--dry-run":
			dryRun = true
		default:
			query = append(query, args[i])
		}
	}
	prompt, err := buildPrompt(query)
	if err != nil {
		return err
	}
	output, completed, err := runAgent(prompt, dryRun)
	if err != nil {
		fmt.Fprintf(a.env.Stderr, "Could not use OpenCode automatically: %v\n", err)
		fmt.Fprintln(a.env.Stderr, "Falling back to a prompt you can paste into any web AI.")
		fmt.Fprintln(a.env.Stderr)
		fmt.Fprintln(a.env.Stdout, prompt)
		return nil
	}
	if completed {
		return a.handleCompletedReport(strings.Join(query, " "), output)
	}
	return nil
}

func handleCompletedReport(query string, output string) error {
	return New("").handleCompletedReport(query, output)
}

func (a *App) handleCompletedReport(query string, output string) error {
	text := textFor(responseLanguage(query))
	for {
		action, value, err := askReportAction(text)
		if err != nil {
			return err
		}
		switch action {
		case reportActionNone:
			return nil
		case reportActionSave:
			path, err := a.saveReport(query, output)
			if err != nil {
				fmt.Fprintf(a.env.Stderr, "Could not save report: %v\n", err)
				return nil
			}
			fmt.Fprintf(a.env.Stderr, "\n%s %s\n", text.SavedReport, path)
			return nil
		case reportActionSkip:
			fmt.Fprintln(a.env.Stderr, text.SkippedSave)
			return nil
		case reportActionDetail:
			prompt, err := buildDetailedPrompt(query, output)
			if err != nil {
				return err
			}
			detail, completed, err := runAgent(prompt, false)
			if err != nil {
				fmt.Fprintf(a.env.Stderr, "Could not use OpenCode automatically: %v\n", err)
				return nil
			}
			if !completed {
				return nil
			}
			output = strings.TrimSpace(output) + "\n\n" + strings.TrimSpace(detail)
		case reportActionNewQuestion:
			return a.runQuery([]string{value})
		}
	}
}

func buildPrompt(queryParts []string) (string, error) {
	system, err := readEmbedded("system-prompt.md")
	if err != nil {
		return "", err
	}
	template, err := readEmbedded("templates/investment-report.md")
	if err != nil {
		return "", err
	}
	workflows, err := readWorkflows()
	if err != nil {
		return "", err
	}
	workflow := `# gacha auto

Classify the user's request into discover, select, entry, exit, portfolio, or journal, then follow the matching gacha workflow. If the request is ambiguous, choose the safest interpretation and state your assumption.`
	query := strings.TrimSpace(strings.Join(queryParts, " "))
	if query == "" {
		query = "(No additional user request supplied.)"
	}
	lang := responseLanguage(query)
	config, err := loadGachaConfig()
	if err != nil {
		return "", fmt.Errorf("could not load research profile config: %w", err)
	}
	sections := []string{
		strings.TrimSpace(system),
		strings.TrimSpace(workflow),
		"Workflow library:\n" + strings.TrimSpace(workflows),
	}
	if profileBlock := profilePromptBlock(config, lang); strings.TrimSpace(profileBlock) != "" {
		sections = append(sections, profileBlock)
	}
	sections = append(sections,
		"User request:\n"+query,
		"Response language:\nWrite the final report in "+string(lang)+". Keep source names, ticker symbols, numbers, and URLs unchanged.",
		"Report template:\n"+strings.TrimSpace(template),
		`Report structure contract:
- The final answer must start with the Easy Basic Report from the template.
- Write for ordinary users, not investment professionals. Use short sentences and explain necessary jargon in plain language.
- The Basic Report is decision-ready: it must include the conclusion, immediate plan, time horizon, action trigger, thesis-break trigger, review timing, main risks, and data freshness.
- The Detailed Analysis is the verification layer: it explains the evidence, valuation, scenarios, portfolio fit, source log, and why the Basic Report's decision rules are reasonable.
- Include Detailed Analysis only when the user asks for it, the request compares multiple choices, the data is mixed, the risk is high, or the recommendation depends on valuation, scenarios, or portfolio fit.
- If Detailed Analysis is included, use the detailed headings from the template in order.
- Always include More Detail Options in the Basic Report so the user knows what they can ask for next.
- Use simple tables only when they make the answer easier to compare, such as candidates, price zones, or action triggers.
- Translate user-facing section headings and labels into the response language, but preserve the template's meaning.`,
		`Hard requirements:
- Always use current web search or current market-data tools before analysis, even if the user does not ask for latest/current/recent data.
- If fresh data cannot be verified, do not make a recommendation.
- Include data freshness, source links, risks, action conditions, and what to monitor next. Include the strongest opposite view when making a recommendation.
- Write user-facing explanations, headings, and action conditions in the response language above.
- Do not execute trades. The final decision remains with the user.`,
	)
	return strings.Join(sections, "\n\n"), nil
}

func buildDetailedPrompt(query string, basicReport string) (string, error) {
	lang := responseLanguage(query + "\n" + basicReport)
	sections := []string{
		"# gacha detailed analysis",
		"User request:\n" + strings.TrimSpace(query),
		"Existing basic report:\n" + strings.TrimSpace(basicReport),
		"Response language:\nWrite the detailed analysis in " + string(lang) + ". Keep source names, ticker symbols, numbers, and URLs unchanged.",
		`Task:
- Continue from the existing basic report.
- Produce only the "Detailed Analysis" section, not a new Basic Report.
- Use these headings in order: Evidence and Sources, Valuation and Scenarios, Strongest Opposite View, Portfolio Fit, Action Rules, Unknowns and Questions, Source Log.
- Use current web search or current market-data tools again when detailed evidence, prices, filings, news, valuation, or source-level claims need verification.
- If fresh data cannot be verified, clearly mark what is missing instead of inventing it.
- Keep explanations plain enough for ordinary users, but include the deeper evidence requested by this detailed view.
- Do not execute trades. The final decision remains with the user.`,
	}
	return strings.Join(sections, "\n\n"), nil
}

func readWorkflows() (string, error) {
	entries, err := fs.ReadDir(agent.FS, "workflows")
	if err != nil {
		return "", err
	}
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		content, err := readEmbedded("workflows/" + entry.Name())
		if err != nil {
			return "", err
		}
		parts = append(parts, strings.TrimSpace(content))
	}
	return strings.Join(parts, "\n\n"), nil
}

func readEmbedded(name string) (string, error) {
	data, err := fs.ReadFile(agent.FS, name)
	if err != nil {
		return "", err
	}
	return string(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))), nil
}
