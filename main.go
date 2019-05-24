package main

import (
	"flag"
	"log"
	"net/url"

	"github.com/michaljemala/expn/pkg/downloader"
	"github.com/michaljemala/expn/pkg/scraper"
)

var (
	flagURL = flag.String("url", "https://www.google.com", "A website URL to be scraped")
	flagDir = flag.String("dir", ".", "Destination directory where website assets will be scraped")
)

func main() {
	flag.Parse()

	URL, err := url.Parse(*flagURL)
	if err != nil {
		log.Fatal("invalid url")
	}

	d, err := downloader.New(
		downloader.WithDestDir(*flagDir),
		downloader.WithConcurrency(10),
	)
	if err != nil {
		log.Fatal(err)
	}

	s := scraper.NewScraper()
	s.RegisterCallback("img[src]", func(e *scraper.HTMLElement) {
		if url := e.AttrValue("src", true); url != "" {
			d.Queue(url)
		}
	})
	s.Scrape(URL)

	d.Stop()
}

