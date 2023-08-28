package assistant

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"github.com/mr-joshcrane/oracle"
)

type Assistant struct {
	oracle *oracle.Oracle
	input  io.Reader
	output io.Writer
}

func Start() error {
	o := oracle.NewOracle()
	assistant := &Assistant{
		oracle: o,
		input:  os.Stdin,
		output: os.Stdout,
	}
	scan := bufio.NewScanner(assistant.input)
	assistant.Prompt("Hello, I am your assistant. How can I help you today?")
	for scan.Scan() {
		line := scan.Text()
		if line == "exit" {
			break
		}
		answer, err := o.Ask(line)
		if err != nil {
			return err
		}
		assistant.Prompt(answer)
		assistant.oracle.GiveExample(line, answer)
	}
	assistant.Prompt("Goodbye!")
	return nil
}

func (a *Assistant) Prompt(prompt string) {
	fmt.Fprint(a.output, "ASSISTANT) ", prompt, "\nUSER) ")
}
