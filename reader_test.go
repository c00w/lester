package lester

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

type testexec struct {
	in  []string
	out chan []byte
	err error
}

func (t *testexec) Run(command ...string) ([]byte, error) {
	t.in = command
	return <-t.out, nil
}

func TestNormalOperation(t *testing.T) {
	m := &testexec{out: make(chan []byte)}
	r := &Reader{
		command:  []string{"command"},
		Incoming: make(chan Message),
		runner:   m,
		l:        sync.Mutex{},
	}
	go r.read()
	defer r.Stop()
	b, err := json.Marshal(rawMessage{
		Envelope: rawEnvelope{
			DataMessage: &rawDataMessage{
				Message: "foo",
			},
		},
	})
	if err != nil {
		t.Fatalf("error marshaling testmessage %v", err)
	}
	m.out <- b

	select {
	case i := <-r.Incoming:
		if i.Body != "foo" {
			t.Errorf("Incorrect body want foo got %s", i.Body)
		}
	case <-time.After(1 * time.Second):
		t.Fatalf("No Message Received")
	}

}
