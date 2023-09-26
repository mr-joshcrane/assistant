package assistant

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/mr-joshcrane/oracle"
)

type Assistant struct {
	Oracle   *oracle.Oracle
	Input    io.Reader
	Output   io.Writer
	AuditLog io.Writer
	History  []QA
}

type QA struct {
	Index    int
	Question string
	Answer   string
}

type Options func(*Assistant) *Assistant

func WithAuditLog(auditLog io.Writer) Options {
	return func(a *Assistant) *Assistant {
		a.AuditLog = auditLog
		a.Input = io.TeeReader(a.Input, auditLog)
		a.Output = NewTeeWriter(a.Output, auditLog)
		return a
	}
}

func WithInput(input io.Reader) Options {
	return func(a *Assistant) *Assistant {
		a.Input = input
		return a
	}
}

func WithOutput(output io.Writer) Options {
	return func(a *Assistant) *Assistant {
		a.Output = output
		return a
	}
}

func NewAssistant(token string, opts ...Options) *Assistant {
	a := &Assistant{
		Oracle: oracle.NewOracle(token),
	}
	for _, opt := range opts {
		a = opt(a)
	}
	if a.Output == nil {
		a.Output = os.Stdout
	}
	if a.Input == nil {
		a.Input = os.Stdin
	}
	if a.AuditLog == nil {
		auditLog, err := CreateAuditLog()
		if err != nil {
			panic(err)
		}
		a.AuditLog = auditLog
		fmt.Println(auditLog.Name())
	}
	a.Input = io.TeeReader(a.Input, a.AuditLog)
	a.Output = NewTeeWriter(a.Output, a.AuditLog)
	return a

}

func (a *Assistant) Start() error {
	scan := bufio.NewScanner(a.Input)
	a.Prompt("Hello, I am your assistant. How can I help you today?")
	for scan.Scan() {
		line := scan.Text()
		err := a.Act(line)
		if err != nil {
			return err
		}
	}
	return scan.Err()
}

func (a *Assistant) Prompt(prompt string) {
	formattedPrompt := fmt.Sprintf("ASSISTANT) %s", prompt)
	color.New(color.FgGreen).Fprintln(a.Output, formattedPrompt)
	color.New(color.FgYellow).Fprint(a.Output, "USER) ")
}

func (a *Assistant) Remember(question string, answer string) {
	if a.History == nil {
		a.History = []QA{}
	}
	a.History = append(a.History, QA{
		Index:    len(a.History) + 1,
		Question: question,
		Answer:   answer,
	})
	a.Oracle.GiveExample(question, answer)
}

func (a *Assistant) Act(line string) error {
	switch {
	case line == "exit":
		return a.Exit()
	case strings.HasPrefix(line, ">"):
		return a.LocalFileSystem(line)
	case strings.HasPrefix(line, "/forget"):
		return a.Forget()
	default:
		return a.Ask(context.Background(), line)
	}
}

func (a *Assistant) Ask(ctx context.Context, question string) error {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan
		cancel()
	}()
	answer, err := a.Oracle.Ask(ctx, question)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			a.Prompt("You cancelled the request, so I'll stop talking!")
			return nil
		}
		return err
	}
	a.Prompt(answer)
	a.Remember(question, answer)
	return nil
}

func (a *Assistant) Exit() error {
	color.New(color.FgGreen).Fprintln(a.Output, "ASSISTANT) Goodbye!")
	return io.EOF
}

func (a *Assistant) LocalFileSystem(line string) error {
	var files []os.DirEntry
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.Contains(path, line[1:]) {
			return nil
		}
		files = append(files, d)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			return err
		}
		a.Remember(string(data), "File Contents of "+path)
		return nil
	})
	allFiles := ""
	for _, file := range files {
		allFiles += "  > " + file.Name() + "\n"
	}
	a.Prompt(fmt.Sprintf("Thanks for the examples!\n%s", allFiles))
	return err
}

func (a *Assistant) Forget() error {
	a.History = []QA{}
	a.Prompt("I've forgotten everything we've talked about!")
	a.Oracle.Reset()
	return nil
}
