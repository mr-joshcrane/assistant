package main

import (
	"fmt"
	"io"
	"os"

	"github.com/mr-joshcrane/assistant"
)

func main() {
	token := os.Getenv("OPENAI_API_KEY")
	if token == "" {
		fmt.Fprintln(os.Stderr, "OPENAI_API_KEY must be set")
		os.Exit(1)
	}
	a := assistant.NewAssistant(token)
	err := a.Start()
	if err != io.EOF {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
