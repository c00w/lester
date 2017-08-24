package lester

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Memory interface {
	SetValue(string, string)
	GetPrefix(prefix string) []BoltPair
}

type FinanceBrain struct {
	W Worlder
	M Memory
}

func IsBalance(m Message) bool {
	return strings.Contains(m.Body, "balance")
}

func IsRich(m Message) bool {
	return strings.Contains(m.Body, "rich")
}

func (e FinanceBrain) CanHandle(m Message) float64 {
	if IsBalance(m) {
		return 1
	}
	if IsRich(m) {
		return 1
	}
	return 0
}

func (e FinanceBrain) Handle(m Message) {
	if IsBalance(m) {
		e.handleBalance(m)
	}
	if IsRich(m) {
		e.handleRich(m)
	}
}

func badInterest(initial, deposit, years int64) int64 {
	i := float64(initial)
	for d := int64(0); d < years; d++ {
		i *= 1.05
		i += float64(deposit)
	}
	return int64(i)
}

func (e FinanceBrain) handleRich(m Message) {
	d := map[string]map[time.Time]int64{}
	for _, e := range e.M.GetPrefix("finance-account") {
		t := strings.Split(e.Key, "-")
		account := t[2]
		t2, _ := strconv.ParseInt(t[4], 10, 64)
		stamp := time.Unix(0, t2)
		balance, _ := strconv.ParseInt(e.Value, 10, 64)
		if d[account] == nil {
			d[account] = map[time.Time]int64{}
		}
		d[account][stamp] = balance
	}
	latest := map[string]time.Time{}
	balance := map[string]int64{}
	for a, m := range d {
		for t, b := range m {
			if t.After(latest[a]) {
				latest[a] = t
				balance[a] = b
			}
		}
	}

	total := int64(0)
	for _, b := range balance {
		total += b
	}

	out := Message{Destination: m.Source, Source: m.Destination}
	out.Body = fmt.Sprintf("You are worth %d$, at 100k/year you'll be worth %d$, 115k %d, 130k %d$",
		total,
		badInterest(total, 100000, 5),
		badInterest(total, 115000, 5),
		badInterest(total, 130000, 5))
	e.W.SendMessage(out)
}

func findAccount(in string) string {
	for _, s := range []string{
		"vanguard",
		"schwab",
		"simple",
	} {
		if strings.Contains(strings.ToLower(in), s) {
			return s
		}

	}
	return ""
}

func findBalance(in string) int64 {
	for _, s := range strings.Split(in, " ") {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return i
		}
	}
	return 0
}
func (e FinanceBrain) handleBalance(m Message) {
	out := Message{Destination: m.Source, Source: m.Destination}
	account := findAccount(m.Body)
	if account == "" {
		out.Body = "unable to parse account what should it have been?"
		e.W.SendMessage(out)
		return
	}
	balance := findBalance(m.Body)
	if balance == 0 {
		out.Body = "unable to parse balance what should it have been?"
		e.W.SendMessage(out)
		return
	}
	e.M.SetValue(fmt.Sprintf("finance-account-%s-time-%d-balance", account, time.Now().UnixNano()), fmt.Sprint(balance))
	out.Body = fmt.Sprintf("Thanks for the info, updated account %s to balance %d", account, balance)
	e.W.SendMessage(out)
}
