package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

type Message struct {
	Destination string
	Body        string
	Source      string
}

type rawMessage struct {
	Envelope struct {
		Source       string
		SourceDevice int
		Timestamp    int64
		IsReceipt    bool
		DataMessage  *struct {
			Timestamp int64
			Message   string
		}
	}
}

type Reader struct {
	command  []string
	incoming chan Message
	l        sync.Mutex
}

func (r *Reader) run(a ...string) ([]byte, error) {
	log.Printf("running %v", a)
	log.Printf("done %v", a)
	r.l.Lock()
	defer r.l.Unlock()
	t := make([]string, 0)
	t = append(t, r.command[1:]...)
	t = append(t, a...)
	return exec.Command(r.command[0], t...).CombinedOutput()
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
			}
			if m.Envelope.DataMessage != nil {
				r.incoming <- Message{Source: m.Envelope.Source, Body: m.Envelope.DataMessage.Message}
			}
		}
	}
}
