package main

import (
	"github.com/mr-joshcrane/assistant"
)

func main() {
	err := assistant.Start()
	if err != nil {
		panic(err)
	}
}
