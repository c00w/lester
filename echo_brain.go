package lester

type EchoBrain struct {
	W Worlder
}

func (e EchoBrain) CanHandle(m Message) float64 {
	return 0.01
}

func (e EchoBrain) Handle(m Message) {
	m.Destination, m.Source = m.Source, m.Destination
	e.W.SendMessage(m)
}
