package lester

import "testing"

type fakememory struct {
	out []BoltPair
}

func (m *fakememory) SetValue(string, string) {}
func (m *fakememory) GetPrefix(prefix string) []BoltPair {
	return nil
}

func TestList(t *testing.T) {
	f := FinanceBrain{M: &fakememory{}}
	if o := f.CanHandle(Message{Body: "what is in my accounts"}); o != 1 {
		t.Errorf("expected %f got %f", 1, o)
	}
}
