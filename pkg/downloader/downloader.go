package downloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/michaljemala/expn/pkg/worker"
)

type Downloader struct {
	concurrency int
	pool        *worker.Pool
	destDir     string
}

type Options func(d *Downloader)

func WithConcurrency(n int) Options {
	return func(d *Downloader) {
		d.concurrency = n
	}
}

func WithDestDir(s string) Options {
	return func(d *Downloader) {
		d.destDir = s
	}
}

func New(opts ...Options) (*Downloader, error) {
	d := &Downloader{
		concurrency: 1,
	}

	for _, opt := range opts {
		opt(d)
	}

	d.pool = worker.NewPool(
		worker.WithConcurrency(d.concurrency),
		worker.WithErrorHandler(func(info interface{}, err error) {
			log.Printf("unable to download %q: %v", info.(string), err)
		}),
	)

	if d.destDir == "" {
		n, err := ioutil.TempDir("", "downloader")
		if err != nil {
			return nil, err
		}
		d.destDir = n
	}

	d.pool.Start()

	return d, nil
}

func (d *Downloader) DestDir() string {
	return d.destDir
}

func (d *Downloader) Queue(url string) {
	d.pool.Submit(worker.Task{
		Info: url,
		Fn: func () error {
			return d.download(url)
		},
	})
}

func (d *Downloader) Stop() {
	d.pool.Stop()
}

func (d *Downloader) download(url string) error {
	filename := path.Base(url)
	if filename == "" {
		return fmt.Errorf("invalid filename")
	}

	filepath := path.Join(d.destDir, filename)

	out, err := os.Create(filepath)
	if err != nil {
		if os.IsExist(err) {
			return nil
		}
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}