package scraper

import (
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

type Config struct {
	Concurrency int
	Client      *http.Client
}

type Scraper interface {
	OnHTMLElement(string, func(*HTMLElement))
	Scrape(*url.URL)
}

type HTMLElement struct {
	n              *html.Node
	resolveURLFunc func(string) (*url.URL, error)
}

func (e HTMLElement) Attr(name string) (string, bool) {
	for _, a := range e.n.Attr {
		if a.Key == name {
			return a.Val, true
		}
	}
	return "", false
}

func (e HTMLElement) ResolveURL(link string) string {
	if e.resolveURLFunc == nil {
		return ""
	}

	u, err := e.resolveURLFunc(link)
	if err != nil {
		return ""
	}

	return u.String()
}
