package assistant_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/mr-joshcrane/assistant"
	"github.com/mr-joshcrane/oracle"
)

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

func TestCreatesAudit_CreatesFileInHomeDirectory(t *testing.T) {
	t.Parallel()
	tDir := t.TempDir()
	original := os.Getenv("HOME")
	defer os.Setenv("HOME", original)
	os.Setenv("HOME", tDir)
	audit, err := assistant.CreateAuditLog()
	if err != nil {
		t.Fatal(err)
	}
	if audit == nil {
		t.Fatal("expected audit log to be created")
	}
	if !strings.Contains(audit.Name(), tDir) {
		t.Fatalf("expected audit log to be created in %s, but it was not", tDir)
	}
}
