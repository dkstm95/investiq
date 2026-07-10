package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type runResultMsg struct {
	query     string
	output    string
	completed bool
	err       error
}

type promptOutputMsg struct {
	query  string
	chunk  string
	stream <-chan tea.Msg
}

type researchPhaseMsg struct{}

type tuiModel struct {
	version string
	lang    language
	text    uiText
	input   textinput.Model
	view    viewport.Model
	spin    spinner.Model
	width   int
	height  int
	busy    bool
	phase   int
	save    *pendingSave
	status  string
	runtime string
	mode    string
	query   string
	report  string
	choice  *pendingChoice
	profile *profileFlow
}

func newTUIModel(version string) tuiModel {
	setThemeStyles(configuredTheme())
	lang := detectLanguage()
	text := textFor(lang)
	input := textinput.New()
	input.Placeholder = text.InputPlaceholder
	input.Focus()
	input.CharLimit = 2000
	input.Prompt = " "

	spin := spinner.New()
	spin.Spinner = spinner.Dot
	spin.Style = accentStyle

	view := viewport.New(80, 16)
	config, configErr := loadGachaConfig()
	profile := (*profileFlow)(nil)
	status := text.Ready
	mode := text.Auto
	if configErr != nil {
		status = text.Fallback
		mode = text.System
		view.SetContent(errorContent(fmt.Errorf("could not load config %s: %w", gachaConfigPath(), configErr), "", text))
	} else if shouldShowProfileOnboarding(config.Profile) {
		profile = newProfileOnboarding(config.Profile)
		view.SetContent(profile.render(lang, 80))
	} else {
		view.SetContent(welcomeContent(version, text, 80, 16))
	}

	return tuiModel{
		version: version,
		lang:    lang,
		text:    text,
		input:   input,
		view:    view,
		spin:    spin,
		width:   100,
		height:  34,
		status:  status,
		runtime: routeLabelFor(lang),
		mode:    mode,
		profile: profile,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.view.Width = max(32, msg.Width-4)
		m.view.Height = max(6, msg.Height-8)
		m.input.Width = max(16, msg.Width-8)
		m.input.Placeholder = inputPlaceholderForWidth(m.text, msg.Width)
		if m.profile != nil {
			m.view.SetContent(m.profile.render(m.lang, m.view.Width))
		} else if m.choice != nil {
			m.view.SetContent(m.choiceContent())
		} else if m.mode == m.text.Auto && !m.busy {
			m.view.SetContent(welcomeContent(m.version, m.text, m.view.Width, m.view.Height))
		} else if m.mode == m.text.Report && m.report != "" && !m.busy {
			if m.save != nil {
				m.view.SetContent(reportContentWithPrompt(m.report, m.text))
			} else {
				m.view.SetContent(renderMarkdownReport(m.report))
			}
		}
	case tea.KeyMsg:
		key := msg.String()
		if m.profile != nil {
			return m.handleProfileKey(key)
		}
		if m.choice != nil {
			switch key {
			case "esc":
				if m.choice.Kind == choiceReport {
					return m.returnToReport(), nil
				}
			case "up", "k":
				m.moveChoice(-1)
				return m, nil
			case "down", "j":
				m.moveChoice(1)
				return m, nil
			}
		}
		switch key {
		case "enter":
			value := strings.TrimSpace(m.input.Value())
			if value == "" && m.choice != nil {
				return m.handleChoiceSelection()
			}
			if value == "" && m.save != nil {
				return m.showReportChoice(), nil
			}
			if value == "" {
				return m, nil
			}
			if m.busy && !isQuitCommand(value) {
				return m, nil
			}
			m.input.SetValue("")
			if m.choice != nil {
				m.choice = nil
			}
			if m.save != nil {
				return m.handleReportAction(value)
			}
			return m.handleSubmit(value)
		}
	case runResultMsg:
		if msg.query != m.query {
			return m, nil
		}
		m.busy = false
		m.phase = 0
		m.runtime = routeLabelFor(m.lang)
		m.mode = m.text.Report
		m.query = msg.query
		if msg.err != nil {
			m.status = m.text.Fallback
			m.report = ""
			m.view.SetContent(errorContent(msg.err, msg.output, m.text))
		} else {
			m.status = m.text.Complete
			if msg.completed && strings.TrimSpace(msg.output) != "" {
				report := strings.TrimSpace(msg.output)
				m.report = report
				m.save = &pendingSave{query: msg.query, report: report}
				m.view.SetContent(reportContentWithPrompt(report, m.text))
			} else {
				m.save = nil
				m.report = strings.TrimSpace(msg.output)
				m.view.SetContent(renderMarkdownReport(msg.output))
			}
		}
		m.view.GotoTop()
	case promptOutputMsg:
		if m.busy && msg.query == m.query {
			m.report += msg.chunk
			clean := displayPromptChunk(m.report)
			if strings.TrimSpace(clean) != "" {
				m.mode = m.text.Report
				m.view.SetContent(strings.TrimSpace(clean))
				m.view.GotoBottom()
			}
		}
		if msg.stream != nil {
			cmds = append(cmds, waitForPromptMsg(msg.stream))
		}
	case researchPhaseMsg:
		if m.busy && len(m.text.ResearchPhases) > 0 {
			m.phase++
			m.status = m.text.ResearchPhases[m.phase%len(m.text.ResearchPhases)]
			cmds = append(cmds, researchPhaseTick())
		}
	}

	if m.busy {
		var cmd tea.Cmd
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)
	}
	var inputCmd tea.Cmd
	m.input, inputCmd = m.input.Update(msg)
	cmds = append(cmds, inputCmd)
	var viewCmd tea.Cmd
	m.view, viewCmd = m.view.Update(msg)
	cmds = append(cmds, viewCmd)
	return m, tea.Batch(cmds...)
}

func (m tuiModel) View() string {
	width := max(44, m.width)
	outerPadding := 4
	if width < 72 {
		outerPadding = 2
	}
	bodyWidth := max(40, width-outerPadding)
	contentHeight := max(6, m.height-7)
	m.view.Width = max(30, bodyWidth-4)
	m.input.Width = max(16, bodyWidth-8)
	m.input.Placeholder = inputPlaceholderForWidth(m.text, m.input.Width)
	workspaceHeight := contentHeight
	m.view.Height = workspaceHeight

	content := m.view.View()
	if m.profile != nil {
		content = m.profile.render(m.lang, m.view.Width)
	} else if m.mode == m.text.Auto && !m.busy {
		content = welcomeContent(m.version, m.text, m.view.Width, workspaceHeight)
		homeHeight := lipgloss.Height(content) + 2
		contentHeight = min(contentHeight, max(8, homeHeight))
		workspaceHeight = contentHeight
		m.view.Height = workspaceHeight
	}

	header := renderHeader(bodyWidth, m.version)
	panel := panelStyle.Width(bodyWidth - 2).Height(contentHeight).Render(content)
	status := renderStatusWithFooter(bodyWidth, m.status, m.runtime, m.mode, m.busy, m.spin.View(), m.text, m.profile == nil)
	parts := []string{header}
	parts = append(parts, panel)
	if m.profile == nil {
		parts = append(parts, renderInput(bodyWidth, m.input.View()))
	}
	parts = append(parts, status)
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func inputPlaceholderForWidth(text uiText, width int) string {
	if width < 72 {
		return text.InputPlaceholderShort
	}
	return text.InputPlaceholder
}

func runPromptCmd(query string) tea.Cmd {
	stream := make(chan tea.Msg, 64)
	return tea.Batch(startPromptCmd(query, stream), waitForPromptMsg(stream))
}

func startPromptCmd(query string, stream chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		result := runPromptWithProgress(query, func(chunk string) {
			stream <- promptOutputMsg{query: query, chunk: chunk, stream: stream}
		})
		stream <- runResultMsg{query: query, output: result.output, completed: result.completed, err: result.err}
		close(stream)
		return nil
	}
}

func runDetailPromptCmd(query string, basicReport string) tea.Cmd {
	stream := make(chan tea.Msg, 64)
	return tea.Batch(startDetailPromptCmd(query, basicReport, stream), waitForPromptMsg(stream))
}

func startDetailPromptCmd(query string, basicReport string, stream chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		result := runDetailedPromptWithProgress(query, basicReport, func(chunk string) {
			stream <- promptOutputMsg{query: query, chunk: chunk, stream: stream}
		})
		stream <- runResultMsg{query: query, output: result.output, completed: result.completed, err: result.err}
		close(stream)
		return nil
	}
}

func waitForPromptMsg(stream <-chan tea.Msg) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-stream
		if !ok {
			return nil
		}
		return msg
	}
}

func displayPromptChunk(chunk string) string {
	clean := stripANSI(chunk)
	clean = strings.ReplaceAll(clean, "\r\n", "\n")
	clean = strings.ReplaceAll(clean, "\r", "\n")
	return clean
}

func researchPhaseTick() tea.Cmd {
	return tea.Tick(2*time.Second, func(time.Time) tea.Msg {
		return researchPhaseMsg{}
	})
}

type promptRunResult struct {
	output    string
	completed bool
	err       error
}

type pendingSave struct {
	query  string
	report string
}
