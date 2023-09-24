package assistant_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"context"

	"github.com/mr-joshcrane/assistant"
	"github.com/mr-joshcrane/oracle"
)

func newTestAssistant(t *testing.T) assistant.Assistant {
	t.Helper()
	o := oracle.NewOracle("", oracle.WithDummyClient("fixed response", 200))
	return assistant.Assistant{
		Oracle: o,
		Input:  new(bytes.Buffer),
		Output: new(bytes.Buffer),
	}
}

func TestAssistant_StartCanBeTerminatedWithExit(t *testing.T) {
	t.Parallel()
	a := newTestAssistant(t)
	a.Input = io.ReadWriter(bytes.NewBufferString("exit\n"))
	err := a.Start()
	if !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
}

func TestAssistant_AsksForUserInput(t *testing.T) {
	t.Parallel()
	a := newTestAssistant(t)
	_ = a.Start()
	a.Input = io.ReadWriter(bytes.NewBufferString("exit\n"))
	got := a.Output.(*bytes.Buffer).String()
	if !strings.Contains(got, "ASSISTANT) ") {
		t.Fatal("expected assistant to ask for user input")
	}
}

func TestAssistant_GivesAResponse(t *testing.T) {
	t.Parallel()
	a := newTestAssistant(t)
	a.Input = io.ReadWriter(bytes.NewBufferString("test\nexit\n"))
	_ = a.Start()
	got := a.Output.(*bytes.Buffer).String()
	if !strings.Contains(got, "fixed response") {
		t.Fatal("expected assistant to give a response")
	}
}

func TestAsk_StopsTalkingWhenRequestCancelled(t *testing.T) {
	t.Parallel()
	a := newTestAssistant(t)
	buf := new(bytes.Buffer)
	a.Output = buf
	a.Oracle = oracle.NewOracle("")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := a.Ask(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(buf)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	if !strings.Contains(got, "cancelled the request") {
		t.Fatalf("expected assistant to stop talking, got %s", got)
	}
}

func TestAuditLog_CapturesAssistantInputOutput(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)

	in := io.ReadWriter(bytes.NewBufferString("test\nexit\n"))
	a := assistant.NewAssistant("", assistant.WithAuditLog(buf), assistant.WithInput(in), assistant.WithOutput(io.Discard))
	a.Oracle = oracle.NewOracle("", oracle.WithDummyClient("fixed output", 200))
	err := a.Start()
	if !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "test\nexit\n") {
		t.Fatalf("expected audit log to contain user input, but it did not. Content: %s", got)
	}
	if !strings.Contains(got, "fixed output") {
		t.Fatalf("expected audit log to contain assistant output, but it did not. Content: %s", got)
	}
}

func TestFile_CreatesAuditLogInXDGDirectory(t *testing.T) {
	t.Parallel()
	tDir := t.TempDir()
	original := os.Getenv("XDG_DATA_HOME")
	defer os.Setenv("XDG_DATA_HOME", original)
	os.Setenv("XDG_DATA_HOME", tDir)
	audit, err := assistant.CreateAuditLog()
	if err != nil {
		t.Fatal(err)
	}
	if audit == nil {
		t.Fatal("expected audit log to be created")
	}
}
