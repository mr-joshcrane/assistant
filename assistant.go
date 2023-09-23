package assistant

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/fatih/color"
	"github.com/mr-joshcrane/oracle"
)

type Assistant struct {
	Oracle  *oracle.Oracle
	Input   io.Reader
	Output  io.Writer
	History []QA
}

type QA struct {
	Index    int
	Question string
	Answer   string
}

func NewAssistant(token string) *Assistant {
	o := oracle.NewOracle(token)
	return &Assistant{
		Oracle: o,
		Input:  os.Stdin,
		Output: os.Stdout,
	}
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
	a.History = append(a.History, QA{
		Index:    len(a.History) + 1,
		Question: question,
		Answer:   answer,
	})
	a.Oracle.GiveExample(question, answer)
}

func (a *Assistant) Act(line string) error {
	switch line {
	case "exit":
		return a.Exit()
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
