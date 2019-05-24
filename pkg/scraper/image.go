package scraper

import (
	"net/url"

	"github.com/gocolly/colly"
)

type Scraper struct {
	concurrency int

	callbacks map[string]func(*HTMLElement)
}

type HTMLElement struct {
	element *colly.HTMLElement
}

func (e HTMLElement) AttrValue(name string, absolute bool) string {
	val := e.element.Attr(name)
	if absolute {
		val = e.element.Request.AbsoluteURL(val)
	}
	return val
}

func NewScraper() *Scraper {
	return &Scraper{
		concurrency: 10,
		callbacks: make(map[string]func(*HTMLElement)),
	}
}

func (s *Scraper) RegisterCallback(selector string, callback func(*HTMLElement)) {
	s.callbacks[selector] = callback
}


func (s *Scraper) Scrape(URL *url.URL) {
	c := colly.NewCollector(
		colly.AllowedDomains(URL.Hostname()),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob: "*",
		Parallelism: s.concurrency,
	})

	for selector, callback := range s.callbacks {
		c.OnHTML(selector, func(e *colly.HTMLElement) {
			if e != nil {
				callback(&HTMLElement{element: e})
			} else {
				callback(nil)
			}
		} )
	}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		url := e.Request.AbsoluteURL(e.Attr("href"))
		if url == "" {
			return
		}

		c.Visit(url)
	})

	c.Visit(URL.String())
	c.Wait()
}
