package lester

import "strings"

type EchoBrain struct {
	W Worlder
}

func (e EchoBrain) CanHandle(m Message) float64 {
	if strings.HasPrefix(m.Body, "echo ") {
		return 1
	}
	return 0.01
}

func (e EchoBrain) Handle(m Message) {
	m.Destination, m.Source = m.Source, m.Destination
	m.Body = strings.TrimLeft(m.Body, "echo ")
	e.W.SendMessage(m)
}
