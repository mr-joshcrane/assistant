package assistant

import (
	"net/http"
	"net/url"

	"github.com/cixtor/readability"
	"github.com/mr-joshcrane/oracle"
)

func TLDR(o *oracle.Oracle, url string) (string, error) {
	o.SetPurpose("Please summarise the provided text as best you can. The shorter the better.")
	content, err := GetContent(url)
	if err != nil {
		return "", err
	}
	return o.Ask(content)
}

// GetContent takes a url, checks if it is a file or URL, and returns the
// contents of the file or the text of the URL.
func GetContent(path string) (msg string, err error) {
	_, err = url.ParseRequestURI(path)
	if err != nil {
		path = "https://" + path
	}
	resp, err := http.Get(path)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	r := readability.New()
	article, err := r.Parse(resp.Body, path)
	if err != nil {
		return "", err
	}
	return article.TextContent, nil
}