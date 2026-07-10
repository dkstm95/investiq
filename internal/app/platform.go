package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const openCodeCommand = "opencode"

type openCodeAuthCredential struct {
	Type string `json:"type"`
}

func runAgent(prompt string, dryRun bool) (string, bool, error) {
	commandPath, ok := resolveCommand(openCodeCommand)
	if !ok || !hasOpenCodeAuth() {
		fmt.Println(prompt)
		return "", false, nil
	}
	resolution := resolveOpenCodeModel()
	args := openCodeRunArgs(prompt, resolution.Model)
	if dryRun {
		parts := []string{commandPath}
		for _, arg := range args {
			parts = append(parts, shellQuote(arg))
		}
		fmt.Println(strings.Join(parts, " "))
		return "", false, nil
	}
	output, err := runOpenCodeWithResolution(commandPath, prompt, resolution, true)
	output = strings.TrimSpace(stripANSI(output))
	return output, err == nil && output != "", err
}

func hasCommand(command string) bool {
	if command == "" {
		return false
	}
	_, ok := resolveCommand(command)
	return ok
}

func hasRunnableCommand(command string) bool {
	commandPath, ok := resolveCommand(command)
	if !ok {
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, commandPath, "--version")
	if err := cmd.Run(); err != nil {
		return false
	}
	return ctx.Err() == nil
}

func resolveCommand(command string) (string, bool) {
	if command == "" {
		return "", false
	}
	if path, err := exec.LookPath(command); err == nil {
		return path, true
	}
	if command == openCodeCommand {
		if path := managedOpenCodePath(); path != "" {
			if _, err := os.Stat(path); err == nil {
				return path, true
			}
		}
	}
	return "", false
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func openCodeRunArgs(prompt string, model string) []string {
	args := []string{"run"}
	if model != "" {
		args = append(args, "--model", model)
	}
	return append(args, prompt)
}

func hasOpenAIChatGPTAuth() bool {
	providers, err := openCodeAuthProviders()
	if err != nil {
		return false
	}
	credential, ok := providers["openai"]
	return ok && strings.EqualFold(credential.Type, "oauth")
}

func openCodeAuthProviders() (map[string]openCodeAuthCredential, error) {
	path := openCodeAuthPath()
	if path == "" {
		return nil, fmt.Errorf("OpenCode auth path not available")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var providers map[string]openCodeAuthCredential
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, err
	}
	return providers, nil
}

func runOpenCodeWithResolution(commandPath string, prompt string, resolution modelResolution, stream bool) (string, error) {
	return runOpenCodeWithResolutionProgress(commandPath, prompt, resolution, stream, nil)
}

func runOpenCodeWithResolutionProgress(commandPath string, prompt string, resolution modelResolution, stream bool, onOutput func(string)) (string, error) {
	candidates := modelCandidates(resolution)
	if len(candidates) == 0 {
		return runOpenCodeWithProgress(commandPath, openCodeRunArgs(prompt, ""), stream, onOutput)
	}

	var lastOutput string
	var lastErr error
	for index, model := range candidates {
		if index > 0 && stream {
			fmt.Fprintf(os.Stderr, "\nOpenCode rejected the selected model. Retrying with %s...\n\n", model)
		}
		output, err := runOpenCodeWithProgress(commandPath, openCodeRunArgs(prompt, model), stream, onOutput)
		if err == nil {
			return output, nil
		}
		lastOutput = output
		lastErr = err
		if !isUnsupportedChatGPTModelError(output) {
			return output, err
		}
	}
	return lastOutput, lastErr
}

func runOpenCode(commandPath string, args []string, stream bool) (string, error) {
	return runOpenCodeWithProgress(commandPath, args, stream, nil)
}

func runOpenCodeWithProgress(commandPath string, args []string, stream bool, onOutput func(string)) (string, error) {
	cmd := exec.Command(commandPath, args...)
	cmd.Stdin = os.Stdin
	if onOutput != nil {
		var out bytes.Buffer
		var outMu sync.Mutex
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return "", err
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			return "", err
		}
		if err := cmd.Start(); err != nil {
			return "", err
		}
		var wg sync.WaitGroup
		copyOutput := func(reader io.Reader) {
			defer wg.Done()
			_, _ = io.Copy(progressWriter{buffer: &out, lock: &outMu, onOutput: onOutput}, reader)
		}
		wg.Add(2)
		go copyOutput(stdout)
		go copyOutput(stderr)
		wg.Wait()
		err = cmd.Wait()
		return out.String(), err
	}
	if stream {
		var out bytes.Buffer
		var outMu sync.Mutex
		cmd.Stdout = captureStreamWriter{
			buffer:   &out,
			lock:     &outMu,
			terminal: &terminalSanitizingWriter{destination: os.Stdout},
		}
		cmd.Stderr = captureStreamWriter{
			buffer:   &out,
			lock:     &outMu,
			terminal: &terminalSanitizingWriter{destination: os.Stderr},
		}
		err := cmd.Run()
		return out.String(), err
	}
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return out.String(), err
}

type progressWriter struct {
	buffer   *bytes.Buffer
	lock     *sync.Mutex
	onOutput func(string)
}

type captureStreamWriter struct {
	buffer   *bytes.Buffer
	lock     *sync.Mutex
	terminal *terminalSanitizingWriter
}

func (w captureStreamWriter) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buffer.Write(p)
	if _, err := w.terminal.Write(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

type terminalSanitizingWriter struct {
	destination io.Writer
	raw         strings.Builder
	emitted     int
}

func (w *terminalSanitizingWriter) Write(p []byte) (int, error) {
	w.raw.Write(p)
	clean := stripANSI(w.raw.String())
	if w.emitted > len(clean) {
		w.emitted = 0
	}
	if _, err := io.WriteString(w.destination, clean[w.emitted:]); err != nil {
		return 0, err
	}
	w.emitted = len(clean)
	return len(p), nil
}

func (w progressWriter) Write(p []byte) (int, error) {
	w.lock.Lock()
	defer w.lock.Unlock()
	w.buffer.Write(p)
	if len(p) > 0 {
		w.onOutput(string(p))
	}
	return len(p), nil
}

func isUnsupportedChatGPTModelError(output string) bool {
	normalized := strings.ToLower(output)
	return strings.Contains(normalized, "model is not supported") &&
		strings.Contains(normalized, "chatgpt account")
}
