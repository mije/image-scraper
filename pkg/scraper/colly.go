package scraper

import (
	"net/url"

	"github.com/gocolly/colly"
)

type collyScraper struct {
	cfg       Config
	callbacks map[string]func(*HTMLElement)
}

func NewCollyScraper(cfg Config) *collyScraper {
	if cfg.Concurrency <= 0 {
		cfg.Concurrency = 10
	}

	return &collyScraper{
		cfg:       cfg,
		callbacks: make(map[string]func(*HTMLElement)),
	}
}

func (s *collyScraper) OnHTMLElement(name string, callback func(*HTMLElement)) {
	s.callbacks[name] = callback
}

func (s *collyScraper) Scrape(URL *url.URL) {
	c := colly.NewCollector(
		colly.AllowedDomains(URL.Hostname()),
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: s.cfg.Concurrency,
	})

	for name, fn := range s.callbacks {
		c.OnHTML(name, func(e *colly.HTMLElement) {
			if e == nil {
				return
			}

			// TODO How we get *html.Node from *colly.HTMLElement
			_ = fn
			//n, err := html.Parse(strings.NewReader(e.Text))
			//if err != nil {
			//	log.Printf("unable to parse node: %s: %v", e.Name, err)
			//	return
			//}
			//
			//fn(&HTMLElement{
			//	n:              n,
			//	resolveURLFunc: e.Request.URL.Parse,
			//})
		})
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
