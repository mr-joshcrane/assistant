package assistant

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/mr-joshcrane/oracle"
)

type Assistant struct {
	Oracle *oracle.Oracle
	Input  io.Reader
	Output io.Writer
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
		if line == "exit" {
			break
		}
		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			<-sigChan
			cancel()
		}()

		answer, err := a.Oracle.Ask(ctx, line)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				a.Prompt("You cancelled the request, so I'll stop talking!")
				continue
			}
			return err
		}
		a.Prompt(answer)
		a.Remember(line, answer)
	}
	a.Prompt("Goodbye!")
	return io.EOF
}

func (a *Assistant) Prompt(prompt string) {
	fmt.Fprint(a.Output, "ASSISTANT) ", prompt, "\nUSER) ")
}

func (a *Assistant) Remember(question string, answer string) {
	a.Oracle.GiveExample(question, answer)
}
