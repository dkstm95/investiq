package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShellQuote(t *testing.T) {
	got := shellQuote("can't")
	if got != "'can'\\''t'" {
		t.Fatalf("unexpected quote: %s", got)
	}
}

func TestOpenCodeRunArgsUsesExplicitModel(t *testing.T) {
	got := openCodeRunArgs("hello", "openai/gpt-5.1-codex")
	want := []string{"run", "--model", "openai/gpt-5.1-codex", "hello"}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("unexpected args: %#v", got)
	}
}

func TestRunOpenCodeWithProgressStreamsOutput(t *testing.T) {
	script := filepath.Join(t.TempDir(), "fake-opencode")
	if err := os.WriteFile(script, []byte("#!/bin/sh\nprintf first\nprintf second >&2\n"), 0o755); err != nil {
		t.Fatal(err)
	}

	var chunks []string
	output, err := runOpenCodeWithProgress(script, nil, false, func(chunk string) {
		chunks = append(chunks, chunk)
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{"first", "second"} {
		if !strings.Contains(output, expected) {
			t.Fatalf("output missing %q: %q", expected, output)
		}
	}
	if strings.Join(chunks, "") == "" {
		t.Fatal("expected progress chunks")
	}
}

func TestTerminalSanitizingWriterHoldsSplitOSCSequence(t *testing.T) {
	var output strings.Builder
	writer := &terminalSanitizingWriter{destination: &output}
	if _, err := writer.Write([]byte("before\x1b]52;c;payload")); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "before" {
		t.Fatalf("incomplete OSC sequence leaked to terminal: %q", got)
	}
	if _, err := writer.Write([]byte("-continued\x07after")); err != nil {
		t.Fatal(err)
	}
	if got := output.String(); got != "beforeafter" {
		t.Fatalf("OSC sequence was not removed: %q", got)
	}
}
