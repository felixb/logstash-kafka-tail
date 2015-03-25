package main

import (
	"os"
	"os/signal"
	"sync"
)

type Printer interface {
	Print(m *Message)
}

type AsyncPrinter struct {
	printer Printer
	wg      sync.WaitGroup
	ch      chan *Message
}

// create a new async printer
func NewAsyncPrinter(p Printer) AsyncPrinter {
	return AsyncPrinter{p, sync.WaitGroup{}, make(chan *Message, 10)}
}

// start the consumer and listen to topic
func (p *AsyncPrinter) Start() {
	// spawn the printer
	p.wg.Add(1)
	go p.printMessages()
}

func (p *AsyncPrinter) Wait() {
	p.wg.Wait()
}

func (p *AsyncPrinter) Print(m *Message) {
	p.ch <- m
}

// reads and filters Messages from chan and prints them to stdout
func (p *AsyncPrinter) printMessages() {
	defer p.wg.Done()

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)

	for {
		select {
		case m := <-p.ch:
			p.printer.Print(m)
		case <-signals:
			return
		}
	}
}
