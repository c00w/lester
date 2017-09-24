package lester

import (
	"strings"
	"testing"
)

type mrunner struct {
	in  []byte
	out []byte
	err error
}

func (m *mrunner) Run(in []byte) ([]byte, error) {
	m.in = in
	return m.out, m.err
}

func TestSend(t *testing.T) {
	m := &mrunner{}
	k := NewKeybaseReader()
	k.run = m
	k.SendMessage(Message{
		Body: "foo",
	})
	if !strings.Contains(string(m.in), "foo") {
		t.Errorf("Unable to find foo in output: %s", string(m.in))
	}
}

func TestReceiveNoMessage(t *testing.T) {
	m := &mrunner{}
	m.out = []byte(`{"result":{"messages":[],"ratelimits":[{"tank":"chat","capacity":900,"reset":165,"gas":827}]}}`)
	k := NewKeybaseReader()
	k.run = m
	_, err := k.ReadMessage()
	if err != NoMessage {
		t.Errorf("Incorrect error received want %v got %v", NoMessage, err)
	}
}

func TestReceive(t *testing.T) {
	m := &mrunner{}
	m.out = []byte(`{"result":{"messages":[{"msg":{"id":15,"channel":{"name":"c00w,lesterbot","public":false,"members_type":"kbfs","topic_type":"chat"},"sender":{"uid":"d377f253073af513966ce1bb3f193d00","username":"c00w","device_id":"8d72525e5bfc9b74303ab08d76195618","device_name":"pixel"},"sent_at":1506216918,"sent_at_ms":1506216918831,"content":{"type":"text","text":{"body":"foo"}},"prev":[{"id":14,"hash":"i4Xg/SQwRPemaHhC0PHPaQ+F9nUI38Xl6Kox0YMIi8I="}],"unread":true}}],"ratelimits":[{"tank":"chat","capacity":900,"reset":472,"gas":857}]}}`)
	k := NewKeybaseReader()
	k.run = m
	b, err := k.ReadMessage()
	if err != nil {
		t.Errorf("Incorrect error received want nil got %v", err)
	}
	if !strings.Contains(b.Body, "foo") {
		t.Errorf("Unable to find foo in output: %s", b.Body)
	}
}
