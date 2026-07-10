package app

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (a *App) startSession() error {
	if !isInteractiveTerminal() {
		return a.startLineSession()
	}
	restoreInterrupts := ignoreTUIInterrupts()
	defer restoreInterrupts()
	program := tea.NewProgram(newTUIModel(a.version), tea.WithAltScreen(), tea.WithoutSignalHandler())
	_, err := program.Run()
	return err
}

func ignoreTUIInterrupts() func() {
	signals := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signals, os.Interrupt)
	go func() {
		for {
			select {
			case <-done:
				return
			case <-signals:
			}
		}
	}()
	return func() {
		signal.Stop(signals)
		close(done)
	}
}

func (a *App) startLineSession() error {
	text := textFor(detectLanguage())
	reader := bufio.NewReader(a.env.Stdin)
	fmt.Fprintln(a.env.Stdout, "gacha")
	if detectLanguage() == languageKorean {
		fmt.Fprintln(a.env.Stdout, "질문을 입력하세요. 종료하려면 /quit을 입력하세요.")
	} else {
		fmt.Fprintln(a.env.Stdout, "Type a question, or /quit to exit.")
	}
	for {
		fmt.Fprint(a.env.Stdout, "\n"+strings.TrimSuffix(text.InputPlaceholder, "...")+" > ")
		line, err := reader.ReadString('\n')
		if err != nil && strings.TrimSpace(line) == "" {
			fmt.Fprintln(a.env.Stdout)
			return nil
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		if isSettingsCommand(input) {
			fmt.Fprintln(a.env.Stdout, settingsContent(text))
			continue
		}
		switch input {
		case "/q", "/quit", "quit", "exit":
			if detectLanguage() == languageKorean {
				fmt.Fprintln(a.env.Stdout, "종료합니다.")
			} else {
				fmt.Fprintln(a.env.Stdout, "Goodbye.")
			}
			return nil
		case "/h", "/help", "help":
			fmt.Fprintln(a.env.Stdout, helpContent(text))
		case "/home", "home":
			fmt.Fprintln(a.env.Stdout, "gacha")
		case "/theme", "theme", "/themes", "themes":
			fmt.Fprintln(a.env.Stdout, settingsContent(text))
		case "/profile", "profile":
			if err := printProfileTo(a.env.Stdout); err != nil {
				fmt.Fprintln(a.env.Stderr, err)
			}
		default:
			if isSlashCommand(input) {
				fmt.Fprintln(a.env.Stdout, fmt.Sprintf(text.UnknownCommand, input))
				fmt.Fprintln(a.env.Stdout)
				fmt.Fprintln(a.env.Stdout, helpContent(text))
				continue
			}
			if err := a.runQuery([]string{input}); err != nil {
				fmt.Fprintln(a.env.Stderr, err)
			}
		}
	}
}

func isInteractiveTerminal() bool {
	return isTerminalFile(os.Stdin) && isTerminalFile(os.Stdout)
}

func isTerminalFile(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
