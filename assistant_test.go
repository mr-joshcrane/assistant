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

func newTestAssistant(t *testing.T, opts ...assistant.Options) *assistant.Assistant {
	t.Helper()
	a := &assistant.Assistant{
		Output:   io.Discard,
		AuditLog: io.Discard,
	}
	o := oracle.NewOracle("", oracle.WithDummyClient("fixed response", 200))
	a.Oracle = o
	for _, opt := range opts {
		a = opt(a)
	}
	return a
}

func TestAssistant_StartCanBeTerminatedWithExit(t *testing.T) {
	t.Parallel()
	a := newTestAssistant(t, assistant.WithInput(io.ReadWriter(bytes.NewBufferString("exit\n"))))
	err := a.Start()
	if !errors.Is(err, io.EOF) {
		t.Fatal(err)
	}
}

func TestAssistant_AsksForUserInput(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	in := assistant.WithInput(io.ReadWriter(bytes.NewBufferString("exit\n")))
	out := assistant.WithOutput(buf)
	a := newTestAssistant(t, in, out)
	_ = a.Start()
	got := buf.String()
	if !strings.Contains(got, "ASSISTANT) ") {
		t.Fatal("expected assistant to ask for user input")
	}
}

func TestAssistant_GivesAResponse(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	in := assistant.WithInput(io.ReadWriter(bytes.NewBufferString("test\nexit\n")))
	out := assistant.WithOutput(buf)
	a := newTestAssistant(t, in, out)
	_ = a.Start()
	got := buf.String()
	if !strings.Contains(got, "fixed response") {
		t.Fatal("expected assistant to give a response")
	}
}

func TestAsk_StopsTalkingWhenRequestCancelled(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	out := assistant.WithOutput(buf)
	a := newTestAssistant(t, out)
	o := oracle.NewOracle("")
	a.Oracle = o
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

func TestAssistant_CanEmbedLocalFiles(t *testing.T) {
	t.Parallel()

	tdir := t.TempDir()
	err := os.WriteFile(tdir+"/file.txt", []byte("embedded content"), 0644)
	if err != nil {
		t.Fatal(err)
	}
	embedCmd := ">" + tdir + "/file.txt"

	buf := new(bytes.Buffer)
	in := assistant.WithInput(io.ReadWriter(bytes.NewBufferString(embedCmd + "\nexit\n")))
	out := assistant.WithOutput(io.Discard)
	audit := assistant.WithAuditLog(buf)
	a := newTestAssistant(t, in, out, audit)
	err = a.Start()
	if err != io.EOF {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "embedded content") {
		t.Fatalf("expected assistant to embed local file, got %s", got)
	}
}
