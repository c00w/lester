package lester

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
)

type Worlder interface {
	SendMessage(m Message) error
	ReadMessage() (Message, error)
}

type Brains interface {
	CanHandle(m Message) float64
	Handle(m Message)
}

type Handler struct {
	quit   chan struct{}
	w      Worlder
	wg     sync.WaitGroup
	cw     clockwork.Clock
	brains []Brains
}

func NewHandler(w Worlder, b ...Brains) *Handler {
	return newHandler(w, clockwork.NewRealClock(), b...)
}

func newHandler(w Worlder, cw clockwork.Clock, b ...Brains) *Handler {
	h := &Handler{quit: make(chan struct{}), w: w, cw: cw, brains: b}
	go h.handle()
	h.wg.Add(1)
	return h
}

func (h *Handler) Close() {
	close(h.quit)
	h.wg.Wait()
}

func (h *Handler) handle() {
	o := time.Duration(1)
	for {
		a := h.cw.After(o)
		o = time.Duration(1)
		select {
		case <-h.quit:
			h.wg.Done()
			return
		case <-a:
		}
		v, err := h.w.ReadMessage()
		if err == NoMessage {
			log.Printf("No messages sleeping")
			o = 10 * time.Second
			continue
		}
		if err != nil {
			log.Printf("Error trying to read message: %v", err)
			continue
		}
		m := 0.0
		mb := h.brains[0]
		for _, b := range h.brains {
			n := b.CanHandle(v)
			fmt.Print(b, v, n)
			if n := b.CanHandle(v); n > m {
				m = n
				mb = b
			}
		}
		mb.Handle(v)
	}
}
