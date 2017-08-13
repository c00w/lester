package lester

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

type CommandRunner interface {
	Run(command ...string) ([]byte, error)
}

type wrapexec struct{}

func (w wrapexec) Run(command ...string) ([]byte, error) {
	return exec.Command(command[0], command[1:]...).CombinedOutput()
}

type Message struct {
	Destination string
	Body        string
	Source      string
}

type rawDataMessage struct {
	Timestamp int64
	Message   string
}
type rawEnvelope struct {
	Source       string
	SourceDevice int
	Timestamp    int64
	IsReceipt    bool
	DataMessage  *rawDataMessage
}
type rawMessage struct {
	Envelope rawEnvelope
}

type Reader struct {
	command  []string
	Incoming chan Message
	runner   CommandRunner
	l        sync.Mutex
	exit     bool
}

func NewReader(command ...string) *Reader {
	r := &Reader{
		command:  command,
		Incoming: make(chan Message),
		runner:   wrapexec{},
		l:        sync.Mutex{},
	}
	go r.read()
	return r
}

func (r *Reader) Stop() {
	r.l.Lock()
	r.exit = true
	r.l.Unlock()
}

func (r *Reader) run(a ...string) ([]byte, error) {
	log.Printf("running %v", a)
	defer log.Printf("done %v", a)
	r.l.Lock()
	defer r.l.Unlock()
	t := make([]string, 0)
	t = append(t, r.command...)
	t = append(t, a...)
	return r.runner.Run(t...)
}

func (r *Reader) SendMessage(m Message) error {
	b, err := r.run("send", m.Destination, "-m", m.Body)
	log.Printf("out %v", string(b))
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}
	return nil
}

func (r *Reader) read() {
	for {

		r.l.Lock()
		if r.exit {
			log.Print("exiting")
			r.l.Unlock()
			return
		}
		r.l.Unlock()
		log.Print("trying receive")
		b, err := r.run("receive", "--json", "--ignore-attachments", "-t", "1")
		if err != nil {
			log.Printf("error receiving message: %v", err)
			continue
		}
		log.Print("received")
		if len(b) == 0 {
			log.Print("no message sleeping")
			time.Sleep(10 * time.Second)
			return
		}
		for d := json.NewDecoder(bytes.NewBuffer(b)); d.More(); {
			m := &rawMessage{}
			if err := d.Decode(m); err != nil {
				log.Printf("Error decoding message: %v", err)
				break
			}
			if m.Envelope.DataMessage != nil {
				r.Incoming <- Message{Source: m.Envelope.Source, Body: m.Envelope.DataMessage.Message}
			}
		}
	}
}
