package lester

import (
	"encoding/json"
	"log"
)

type KeybaseReader struct {
	store []Message
	run   CommandRunner
}

type kMessage struct {
	Body string `json:"body"`
}
type kCommand struct {
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
	Result struct {
		Messages []struct {
			Msg struct {
				Content struct {
					Type string
					Text struct {
						Body string
					}
				}
			}
		}
	}
}

func NewKeybaseReader() *KeybaseReader {
	return &KeybaseReader{nil, wrapexec{}}
}

func (k *KeybaseReader) SendMessage(m Message) error {
	com := kCommand{Method: "send"}
	com.Params.Options.Channel.Name = "lesterbot,c00w"
	com.Params.Options.Message = &kMessage{Body: m.Body}
	b, err := json.Marshal(com)
	if err != nil {
		return err
	}
	log.Printf("running %s", string(b))
	o, err := k.run.Run(b)
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
	read := kCommand{Method: "read"}
	read.Params.Options.Unread_only = true
	read.Params.Options.Channel.Name = "lesterbot,c00w"
	b, err := json.Marshal(read)
	if err != nil {
		return Message{}, err
	}
	log.Printf("running %s", string(b))
	o, err := k.run.Run(b)
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
