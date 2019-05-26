package scraper

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type customScraper struct {
	config    Config
	callbacks map[string]func(*HTMLElement)
}

func NewCustomScraper(cfg Config) *customScraper {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 10
	}
	if cfg.Client == nil {
		cfg.Client = http.DefaultClient
	}
	return &customScraper{
		config:    cfg,
		callbacks: make(map[string]func(*HTMLElement)),
	}
}

func (s *customScraper) OnHTMLElement(name string, fn func(*HTMLElement)) {
	s.callbacks[name] = fn
}

func (s *customScraper) Scrape(u *url.URL) {
	seed := make(chan []string)
	unseen := make(chan string)

	allowedDomain := func(found *url.URL) bool {
		return found.Hostname() == u.Hostname()
	}

	go func() {
		seen := make(map[string]bool)
		for list := range seed {
			for _, link := range list {
				if !seen[link] {
					seen[link] = true
					unseen <- link
				}
			}
		}
	}()

	wg := &sync.WaitGroup{}
	for i := 0; i < s.config.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case <-time.After(1 * time.Second):
					return // We have nothing to read for some time, we end !?!
				case link := <-unseen:
					doc, err := s.getDocument(link)
					if err != nil {
						s.handleError(link, err)
						continue
					}

					resolveURL := func(ref string) (*url.URL, error) {
						return doc.req.URL.Parse(ref)
					}

					var foundLinks []string
					s.walkDocument(doc, func(n *html.Node) {
						if n == nil || n.Type != html.ElementNode {
							return
						}

						// Collect links
						if n.Data == "a" {
							for _, a := range n.Attr {
								if a.Key != "href" {
									continue
								}
								link, err := resolveURL(a.Val)
								if err != nil {
									continue
								}
								if allowedDomain(link) {
									foundLinks = append(foundLinks, link.String())
								}
							}
						}

						// Call registered callbacks
						if fn, ok := s.callbacks[n.Data]; ok {
							fn(&HTMLElement{
								n:              n,
								resolveURLFunc: resolveURL,
							})
						}
					})

					if len(foundLinks) > 0 {
						// Send the links asynchronously as we use unbuffered channel
						go func() { seed <- foundLinks }()
					}
				}
			}
		}()
	}

	go func() { seed <- []string{u.String()} }()

	wg.Wait()
	close(seed)
}

type document struct {
	root *html.Node
	req  *http.Request
}

func (s *customScraper) getDocument(link string) (*document, error) {
	resp, err := s.config.Client.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unable to get document: %s: %s", link, resp.Status)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to parse document: %s: %v", link, err)
	}

	return &document{
		root: doc,
		req:  resp.Request,
	}, nil
}

func (s *customScraper) handleError(link string, err error) {
	log.Printf("unable to load document: %s: %v", link, err)
}

func (s *customScraper) walkDocument(doc *document, fn func(n *html.Node)) {
	if doc == nil || fn == nil {
		return
	}

	walk(doc.root, fn)
}

func walk(n *html.Node, fn func(*html.Node)) {
	fn(n)

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		walk(c, fn)
	}
}
