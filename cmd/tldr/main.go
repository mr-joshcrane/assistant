package main

import (
	"fmt"
	"net/http"

	"github.com/mr-joshcrane/assistant"
)

func main() {
	http.ListenAndServe("localhost:8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		err := r.ParseForm()
		url := r.FormValue("url")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		answer, err := assistant.TLDR(url)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		fmt.Fprintf(w, "%s", answer)
	}))
}
