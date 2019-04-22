package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

const (
	rangeSizeDefault  = int64(10485760)
	bufferSizeDefault = int64(10240)
	rangesDefault     = 40
	workersDefault    = 10
)

func readURLsFile(file string) (urls []string, err error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rd := bufio.NewReader(f)
	for {
		line, err := rd.ReadString('\n')
		if err != nil || io.EOF == err {
			break
		}
		urls = append(urls, strings.TrimSuffix(line, "\n"))
	}
	return urls, nil
}

func Downloads(urls []string, dir string) {
	pb := NewProgressBar()

	var wg sync.WaitGroup
	mch := make(map[int]chan bool)

	semaWorkers := make(chan struct{}, workersDefault)
	for id, url := range urls {

		wg.Add(1)
		saveFile := dir + "/" + path.Base(url)

		dlr, err := NewDownloader(url, saveFile)
		if err != nil {
			log.Fatal(err)
		}
		defer dlr.writer.Close()

		go dlr.Download(semaWorkers)
		compch := make(chan bool)
		mch[id] = compch

		go func(id int, ch chan bool, down *Downloader) {
			pb.Add(id, down)
			ch <- true
		}(id, compch, dlr)
	}

	for _, ch := range mch {
		pb.Wait(&wg, ch)
	}

	wg.Wait()
}

type Rang struct {
	Begin int64
	End   int64
}

func createRanges(totalSize int64, bs int64) (ranges []Rang) {
	var begin int64
	var end int64

	for begin < totalSize {
		end += bs
		ranges = append(ranges, Rang{begin, end})
		begin = end
	}
	ranges[len(ranges)-1].End = totalSize - 1
	return
}

type status struct {
	completing int64
	completed  int64
}

type Downloader struct {
	capacity     int
	ThreadSize   int
	RangeSize    int64
	BufferSize   int64
	totalSize    int64
	ranges       []Rang
	status       *status
	httpClient   *http.Client
	completed    *sync.WaitGroup
	semaRanges   chan struct{}
	finished     chan bool
	showProgress bool

	nch chan int

	filename string
	url      string
	writer   *os.File
}

func NewDownloader(url string, name string) (*Downloader, error) {
	f, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}

	return &Downloader{
		ThreadSize:   rangesDefault,
		RangeSize:    rangeSizeDefault,
		BufferSize:   bufferSizeDefault,
		ranges:       []Rang{},
		status:       &status{},
		httpClient:   &http.Client{},
		completed:    &sync.WaitGroup{},
		semaRanges:   make(chan struct{}, rangesDefault),
		finished:     make(chan bool),
		showProgress: true,
		nch:          make(chan int),

		filename: name,
		url:      url,
		writer:   f,
	}, nil
}

func (d *Downloader) SetTotalSize(n int64) {
	d.totalSize = n
}

func (d *Downloader) SetRanges(ranges []Rang) {
	d.ranges = ranges
}

func (d *Downloader) SetFileName(name string) {
	d.filename = name
}

func (d *Downloader) SetBufferSize(bs int64) {
	d.BufferSize = bs
}

func (d *Downloader) SetThreadSize(ts int) {
	d.ThreadSize = ts
}

func (s *status) Speed(n int64) int64 {
	speed := s.completing / n
	s.completed = s.completing
	return speed
}

func (d *Downloader) Download(sema chan struct{}) {
	sema <- struct{}{}
	defer func() { <-sema }()
	resp, err := d.httpClient.Head(d.url)
	if err != nil {
		log.Fatalln(err)
	}
	header := resp.Header
	clength := header.Get("Content-Length")
	totalSize, err := strconv.ParseInt(clength, 10, 64)
	if err != nil {
		log.Fatalln(err)
	}
	d.SetTotalSize(totalSize)
	acceptRange := header.Get("Accept-Ranges")
	if acceptRange != "bytes" {
		log.Fatalln("the Request http not support accept ranges")
	}

	ranges := createRanges(d.totalSize, d.RangeSize)
	d.SetRanges(ranges)

	for i, _ := range d.ranges {
		d.completed.Add(1)
		go func(id int) {
			if err := d.DownloadRange(id); err != nil {
				d.completed.Add(1)
				d.DownloadRange(id)
			}
		}(i)
	}
	d.completed.Wait()
	d.finished <- true
}

func (d *Downloader) DownloadRange(id int) error {
	defer d.completed.Done()
	d.semaRanges <- struct{}{}
	defer func() { <-d.semaRanges }()

	req, err := http.NewRequest("GET", d.url, nil)
	if err != nil {
		return err
	}

	rang := d.ranges[id]
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", rang.Begin, rang.End))
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	offset := rang.Begin
	p := make([]byte, d.BufferSize)
	for {
		bs, err := resp.Body.Read(p)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		_, err = d.writer.WriteAt(p, offset)
		if err != nil {
			return err
		}
		offset += int64(bs)
		atomic.AddInt64(&d.status.completing, int64(bs))
	}
	return nil
}
