package lester

import (
	"bytes"
	"encoding/json"
	"log"
	"os/exec"
)

type KeybaseReader struct {
	store []Message
}

type kMessage struct {
	Body string `json:"body"`
}
type kcommand struct {
	Method string `json:"method"`
	Params struct {
		Options struct {
			Unread_only bool `json:"unread_only,omitempty"`
			Channel     struct {
				Name string `json:"name"`
			} `json:"channel"`
			Message *kMessage `json:"message,omitempty"`
		} `json:"options"`
	} `json:"params"`
}

type kresponse struct {
	Result kresult
}

type ktext struct {
	Body string
}

type kcontent struct {
	Type string
	Text ktext
}

type kmsg struct {
	Content kcontent
}

type kmessage struct {
	Msg kmsg
}

type kresult struct {
	Messages []kmessage
}

func (k *KeybaseReader) run(input []byte) ([]byte, error) {
	c := exec.Command("keybase", "chat", "api")
	c.Stdin = bytes.NewBuffer(input)
	return c.CombinedOutput()
}

func (k *KeybaseReader) SendMessage(m Message) error {
	com := kcommand{Method: "send"}
	com.Params.Options.Channel.Name = "lesterbot,c00w"
	com.Params.Options.Message = &kMessage{Body: m.Body}
	b, err := json.Marshal(com)
	if err != nil {
		return err
	}
	log.Printf("running %s", string(b))
	o, err := k.run(b)
	if err != nil {
		return err
	}
	log.Printf("received %s", string(o))
	return nil
}

func (k *KeybaseReader) ReadMessage() (Message, error) {
	log.Print("Waiting for message")
	if len(k.store) > 0 {
		out := k.store[0]
		k.store = k.store[1:]
		return out, nil
	}
	read := kcommand{Method: "read"}
	read.Params.Options.Unread_only = true
	read.Params.Options.Channel.Name = "lesterbot,c00w"
	b, err := json.Marshal(read)
	if err != nil {
		return Message{}, err
	}
	log.Printf("running %s", string(b))
	o, err := k.run(b)
	if err != nil {
		return Message{}, err
	}
	log.Printf("received %s", string(o))
	r := &kresponse{}
	if err := json.Unmarshal(o, &r); err != nil {
		return Message{}, err
	}
	log.Printf("decoded %+v", r)

	m := []Message{}
	for _, i := range r.Result.Messages {
		m = append(m, Message{
			Body: i.Msg.Content.Text.Body,
		})
	}
	if len(m) == 0 {
		return Message{}, NoMessage
	}
	if len(m) > 1 {
		k.store = m[1:len(m)]
	}
	return m[0], nil
}
