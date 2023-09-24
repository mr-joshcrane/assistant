package assistant

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// CreateAuditLog creates a new file for recording the conversation's audit
// log with a timestamped filename. This function helps ensure
// conversation logs are saved and timestamped, allowing the users to
// review their chat history later.
func CreateAuditLog() (*os.File, error) {
	auditLogDir, err := getAuditLogDir()
	if err != nil {
		return nil, err
	}
	dateTimeString := time.Now().Format("2006-01-02_15-04-05")
	return os.Create(filepath.Join(auditLogDir, fmt.Sprintf("%s.log", dateTimeString)))
}

func getAuditLogDir() (string, error) {
	// Use XDG_STATE_HOME if available, otherwise fallback to default
	xdgStateHome := os.Getenv("XDG_STATE_HOME")
	if xdgStateHome == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		xdgStateHome = filepath.Join(home, ".local", "state")
	}

	// Create your application's specific directory for storing audit logs
	appAuditLogDir := filepath.Join(xdgStateHome, "assistant", "audit_logs")
	err := os.MkdirAll(appAuditLogDir, 0700)
	if err != nil {
		return "", err
	}

	return appAuditLogDir, nil
}

type TeeWriter struct {
	writers []io.Writer
}

func NewTeeWriter(writers ...io.Writer) io.Writer {
	var combinedWriters []io.Writer
	combinedWriters = append(combinedWriters, writers...)
	return TeeWriter{writers: combinedWriters}
}

func (t TeeWriter) Write(p []byte) (n int, err error) {
	for _, w := range t.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return n, nil
}
