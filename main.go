package main

import (
	"flag"
	"log"
	"net/url"
	"os"

	"github.com/michaljemala/image-scraper/pkg/downloader"
	"github.com/michaljemala/image-scraper/pkg/scraper"
)

var (
	flagURL = flag.String("url", "https://exponea.com", "A website URL to be scraped")
	flagDir = flag.String("dir", "./data", "Destination directory where website assets will be scraped")
)

func main() {
	flag.Parse()

	u, err := url.Parse(*flagURL)
	if err != nil {
		log.Fatal("invalid url")
	}

	if err := os.MkdirAll(*flagDir, os.ModePerm); err != nil {
		log.Fatal("unable to create dir")
	}

	d, err := downloader.New(
		downloader.WithDestDir(*flagDir),
		downloader.WithConcurrency(10),
	)
	if err != nil {
		log.Fatal(err)
	}

	s := scraper.NewCustomScraper(scraper.Config{
		Concurrency: 10,
	})
	s.OnHTMLElement("img", func(e *scraper.HTMLElement) {
		link, ok := e.Attr("src")
		if !ok {
			return
		}
		link = e.ResolveURL(link)
		if link != "" {
			log.Printf("enqueueing: %s", link)
			d.Queue(link)
		}
	})
	s.Scrape(u)

	d.Stop()
}
