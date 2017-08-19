package lester

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Memory interface {
	SetValue(string, string)
}

type FinanceBrain struct {
	W Worlder
	M Memory
}

func (e FinanceBrain) CanHandle(m Message) float64 {
	if strings.Contains(m.Body, "balance") {
		return 1
	}
	return 0
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

func (e FinanceBrain) Handle(m Message) {
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
