package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mr-joshcrane/assistant"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("Request: %v\n", r)
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		url := r.FormValue("summaryUrl")

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		fmt.Println("url: ", url)
		summary, err := assistant.TLDR(url)
		if err != nil {
			fmt.Println(err)
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
		fmt.Fprintf(w, "%s", summary)
	}))
	mux.HandleFunc("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "cmd/tldr/index.html")
	}))
	log.Fatal(http.ListenAndServe("localhost:8082", mux))
}

func handleX() {
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
