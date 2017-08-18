package lester

import (
	"log"
	"sync"
	"time"

	"github.com/jonboulle/clockwork"
)

type Worlder interface {
	SendMessage(m Message) error
	ReadMessage() (Message, error)
}

type Handler struct {
	quit chan struct{}
	w    Worlder
	wg   sync.WaitGroup
	cw   clockwork.Clock
}

func NewHandler(w Worlder) *Handler {
	return newHandler(w, clockwork.NewRealClock())
}

func newHandler(w Worlder, cw clockwork.Clock) *Handler {
	h := &Handler{quit: make(chan struct{}), w: w, cw: cw}
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
		log.Printf("%v %q", v, v.Body)
		v.Destination, v.Source = v.Source, v.Destination
		h.w.SendMessage(v)
	}
}
