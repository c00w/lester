package lester

import (
	"testing"

	"github.com/jonboulle/clockwork"
)

type mockworld struct {
	send Message
	recv Message
}

func (w *mockworld) SendMessage(m Message) error {
	w.recv = m
	return nil
}

func (w *mockworld) ReadMessage() (Message, error) {
	return w.send, nil
}

type textTest struct {
	in  Message
	out Message
}

func TestEcho(t *testing.T) {
	for _, tc := range []textTest{
		{
			in:  Message{Source: "source", Destination: "dest", Body: "message"},
			out: Message{Source: "dest", Destination: "source", Body: "message"},
		},
	} {
		m := &mockworld{send: tc.in}
		c := clockwork.NewFakeClock()
		h := newHandler(m, c)
		defer h.Close()
		c.BlockUntil(1)
		c.Advance(1)
		c.BlockUntil(1)

		if m.recv != tc.out {
			t.Fatalf("Incorrect message, expected %#v, got %#v", tc.out, m.recv)
		}
	}
}
