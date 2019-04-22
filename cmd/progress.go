package main

import (
	"bufio"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/sethgrid/curse"
)

func (p *ProgressBar) Add(id int, d *Downloader) {
	p.bufSpace()
	var speed string
	nt := int64(0)
	for {
		nt++
		speed = BytesToHuman(float64(d.status.Speed(nt))) + "/s"
		select {
		case <-d.finished:
			p.Printf(id, "⬇︎ %s 100%% %d/%d %-15s %-13s\n", d.filename, d.totalSize, d.totalSize, speed, "[Completed]")
			return
		default:
			progress := float64(d.status.completed) / float64(d.totalSize) * 100
			p.Printf(id, "⬇︎ %s %.2f%% %d/%d %-15s %-13s\n", d.filename, progress, d.status.completed, d.totalSize, speed, "[InProgess]")
			time.Sleep(time.Second)
		}
	}
}

func (p *ProgressBar) Start(id int, nch chan int) {
	p.bufSpace()
	for {
		select {
		case completing, ok := <-nch:
			if !ok {
				return
			}
			p.Printf(id, "[%d] progress bar test %d\n", id, completing)
		default:
			time.Sleep(time.Second)
		}
	}
}

func (p *ProgressBar) Wait(wg *sync.WaitGroup, compch chan bool) {
	for {
		select {
		case <-compch:
			wg.Done()
			return
		default:
			time.Sleep(1 * time.Second)
			p.Flush()
		}
	}
}

type ProgressBar struct {
	output *bufio.Writer

	history map[int]string
	mtx     *sync.RWMutex
}

func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		output:  bufio.NewWriter(os.Stdout),
		history: make(map[int]string),
		mtx:     new(sync.RWMutex),
	}
}

func (p *ProgressBar) bufSpace() {
	fmt.Println()
}

func (p *ProgressBar) Flush() {
	c, _ := curse.New()
	total := len(p.history)
	c.MoveUp(total)
	for n := 0; n < total; n++ {
		p.output.WriteString(p.history[n])
	}
	p.output.Flush()
}

func (p *ProgressBar) Print(id int, a ...interface{}) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.history[id] = fmt.Sprint(a...)
	return
}

func (p *ProgressBar) Println(id int, a ...interface{}) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.history[id] = fmt.Sprintln(a...)
	return
}

func (p *ProgressBar) Printf(id int, format string, a ...interface{}) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	p.history[id] = fmt.Sprintf(format, a...)
	return
}
