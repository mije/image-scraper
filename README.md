# Image Scraper

Write a crawler that downloads all images from a goiven website.

It uses the [Colly](https://github.com/gocolly/colly) to traverse website's pages recursivelly and collects image links. Collected links are processed by a custom-build concurrent downloader.

### Downloader
A very simple concurrent file downloader. It misses some features to make it production ready, e.g.:
* Graceful shutdown
* Retrying failed requests
* Rate limiters with custom backoff strategies
