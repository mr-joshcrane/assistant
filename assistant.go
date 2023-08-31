package assistant

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/mr-joshcrane/oracle"
)

type Assistant struct {
	oracle *oracle.Oracle
	input  io.Reader
	output io.Writer
}

func Start() error {
	o := oracle.NewOracle(os.Getenv("OPENAI_API_KEY"))
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
		assistant.Remember(line, answer)
	}
	assistant.Prompt("Goodbye!")
	return nil
}

func (a *Assistant) Prompt(prompt string) {
	fmt.Fprint(a.output, "ASSISTANT) ", prompt, "\nUSER) ")
}

func (a *Assistant) Remember(question string, answer string) {
	a.oracle.GiveExample(question, answer)
}

type TLDRServer struct {
	oracle     *oracle.Oracle
	httpServer *http.Server
}

func NewTLDRServer(o *oracle.Oracle, addr string) *TLDRServer {
	s := &TLDRServer{
		oracle: o,
		httpServer: &http.Server{
			Addr: addr,
		},
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat/", s.chatHandler)
	mux.HandleFunc("/", s.indexHandler)

	s.httpServer.Handler = mux
	return s
}

func (s *TLDRServer) indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TLDR")
}

func (s *TLDRServer) chatHandler(w http.ResponseWriter, r *http.Request) {
	req, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println(string(req))
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	defer r.Body.Close()
	err = r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
	url := r.FormValue("summaryUrl")

	fmt.Println("url: ", url)
	summary, err := TLDR(s.oracle, url)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
	htmlFragment := fmt.Sprintf("<div class='results'><h3><a href='%s' target='_blank'>%s</a></h3>%s</div>", url, url, summary)

	fmt.Fprintf(w, "%s", htmlFragment)

}

func (s *TLDRServer) ListenAndServe() error {
	return s.httpServer.ListenAndServe()
}

func (s *TLDRServer) Shutdown() error {
	return s.httpServer.Shutdown(nil)
}

// mux := http.NewServeMux()
