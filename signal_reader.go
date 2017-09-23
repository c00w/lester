package lester

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"sync"
)

var (
	NoMessage = errors.New("No message Received")
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
	Attachments []string
}

type Attachment struct {
	ID          uint64
	Size        uint64
	ContentType string
}

type rawDataMessage struct {
	Timestamp   int64
	Message     string
	Attachments []Attachment
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

type SignalReader struct {
	command []string
	runner  CommandRunner
	l       sync.Mutex
	buf     []*Message
}

func NewSignalReader(command ...string) *SignalReader {
	r := &SignalReader{
		command: command,
		runner:  wrapexec{},
	}
	return r
}

func (r *SignalReader) run(a ...string) ([]byte, error) {
	log.Printf("running %v", a)
	defer log.Printf("done %v", a)
	r.l.Lock()
	defer r.l.Unlock()
	t := make([]string, 0)
	t = append(t, r.command...)
	t = append(t, a...)
	return r.runner.Run(t...)
}

func (r *SignalReader) SendMessage(m Message) error {
	a := []string{"send", m.Destination, "-m", m.Body}
	if len(m.Attachments) > 0 {
		for _, m := range m.Attachments {
			a = append(a, "-a", m)
		}
	}
	b, err := r.run(a...)
	log.Printf("out %v", string(b))
	if err != nil {
		return fmt.Errorf("error sending message: %v", err)
	}
	return nil
}

func (r *SignalReader) ReadMessage() (Message, error) {
	for {
		if len(r.buf) > 0 {
			t := r.buf[0]
			r.buf = r.buf[1:]
			return *t, nil

		}
		log.Print("trying receive")
		b, err := r.run("receive", "--json", "-t", "1")
		if err != nil {
			return Message{}, fmt.Errorf("error receiving messages: %v", err)
		}
		log.Print("received")
		if len(b) == 0 {
			return Message{}, NoMessage
		}
		for d := json.NewDecoder(bytes.NewBuffer(b)); d.More(); {
			m := &rawMessage{}
			if err := d.Decode(m); err != nil {
				return Message{}, fmt.Errorf("error decoding message: %v", err)
			}
			log.Printf("%s", string(b))
			if m.Envelope.DataMessage != nil {
				r.buf = append(r.buf, &Message{Source: m.Envelope.Source, Body: m.Envelope.DataMessage.Message})
				for _, m := range m.Envelope.DataMessage.Attachments {
					fn := fmt.Sprintf("/home/colin/.config/signal/attachments/%d", m.ID)
					r.buf[len(r.buf)-1].Attachments = append(r.buf[len(r.buf)-1].Attachments, fn)
				}
			}
		}
	}
}
