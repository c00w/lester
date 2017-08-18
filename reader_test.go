package lester

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

type testexec struct {
	in  []string
	out []byte
	err error
}

func (t *testexec) Run(command ...string) ([]byte, error) {
	t.in = command
	return t.out, nil
}

func TestNormalOperation(t *testing.T) {
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
	m := &testexec{out: b}
	r := &Reader{
		command:  []string{"command"},
		Incoming: make(chan Message),
		runner:   m,
		l:        sync.Mutex{},
	}

	i, err := r.ReadMessage()
	if err != nil {
		t.Fatalf("Error reading message: %v", err)
		time.Sleep(10)
	}
	if i.Body != "foo" {
		t.Errorf("Incorrect body want foo got %s", i.Body)
	}

}
