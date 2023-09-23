package assistant_test

import (
	"bytes"
	"errors"
	"io"
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
