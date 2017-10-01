package lester

import (
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	chart "github.com/wcharczuk/go-chart"
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

func IsListBalance(m Message) bool {
	return strings.Contains(m.Body, "accounts")
}

func IsRich(m Message) bool {
	return strings.Contains(m.Body, "rich")
}

func IsReport(m Message) bool {
	return strings.Contains(m.Body, "report")
}

func IsCSV(m Message) bool {
	return strings.Contains(m.Body, "csv") && len(m.Attachments) == 1
}

func (e FinanceBrain) CanHandle(m Message) float64 {
	if IsListBalance(m) {
		return 1
	}
	if IsBalance(m) {
		return 1
	}
	if IsRich(m) {
		return 1
	}
	if IsReport(m) {
		return 1
	}
	if IsCSV(m) {
		return 1

	}
	return 0
}

func (e FinanceBrain) Handle(m Message) {
	if IsListBalance(m) {
		e.handleList(m)
		return
	}
	if IsBalance(m) {
		e.handleBalance(m)
		return
	}
	if IsRich(m) {
		e.handleRich(m)
		return
	}
	if IsReport(m) {
		e.handleReport(m)
	}
	if IsCSV(m) {
		e.handleCSV(m)
	}
}

type rawEntry struct {
	account string
	data    time.Time
	delta   int64
}

func (e FinanceBrain) handleCSV(m Message) {
	f, err := os.Open(m.Attachments[0])
	if err != nil {
		log.Fatalf("Error opening file %s: %v", m.Attachments[0], err)
	}

	firstblock := false
	seeksecond := false
	secondblock := false
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	for {
		col, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("error reading row %s", err)
		}

		if !firstblock {
			if len(col) == 0 || col[0] != "Account Number" {
				continue
			}
			firstblock = true
			continue
		}

		if firstblock && !seeksecond {
			if len(col) <= 8 {
				seeksecond = true
				continue
			}
			t, err := time.Parse("01/02/2006", col[2])
			if err != nil {
				log.Fatalf("Error parsing date: %v", err)
			}
			bal, err := strconv.ParseFloat(col[8], 64)
			if err != nil {
				log.Fatalf("Error parsing balance: %v", err)
			}
			log.Print(rawEntry{col[5], t, int64(bal)})

		}
		if seeksecond && !secondblock {
			if len(col) == 0 || col[0] != "Account Number" {
				continue
			}
			secondblock = true
			continue
		}
		if secondblock {
			t, err := time.Parse("01/02/2006", col[1])
			if err != nil {
				log.Fatalf("Error parsing date: %v", err)
			}
			bal, err := strconv.ParseFloat(col[8], 64)
			if err != nil {
				log.Fatalf("Error parsing balance: %v", err)
			}
			log.Print(rawEntry{col[5], t, int64(bal)})
		}
	}

}

func Interest(initial, deposit int64, end time.Time) int64 {
	endd := int64(end.Sub(time.Now()) / (30 * 24 * time.Hour))
	i := float64(initial)
	for d := int64(0); d < endd; d++ {
		i *= math.Exp(0.05 * 1.0 / 12)
		i += float64(deposit) / 12
	}
	return int64(i)
}

func (e FinanceBrain) handleList(m Message) {
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
	mess := "Here are all your accounts:"
	for a, b := range balance {
		mess += fmt.Sprintf(" account %s has balance %d", a, b)
	}
	out := Message{Destination: m.Source, Source: m.Destination}
	out.Body = mess
	e.W.SendMessage(out)
}

func (e FinanceBrain) buildAccounts() map[string]map[time.Time]int64 {
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
	return d
}

func (e FinanceBrain) totalBalance() int64 {
	d := e.buildAccounts()
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
	return total
}

func (e FinanceBrain) handleRich(m Message) {
	total := e.totalBalance()

	bday, _ := time.Parse("2006/01/02", "2022/02/21")
	out := Message{Destination: m.Source, Source: m.Destination}
	out.Body = fmt.Sprintf("You are worth %d$, on your 30th birthday at 100k/year you'll be worth %d$, 115k %d, 130k %d$",
		total,
		Interest(total, 100000, bday),
		Interest(total, 115000, bday),
		Interest(total, 130000, bday))
	e.W.SendMessage(out)
}

func findAccount(in string) string {
	for _, s := range []string{
		"vanguard",
		"schwab",
		"simple",
		"hsa",
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

type sortT struct {
	t []time.Time
	b []float64
}

func (s sortT) Len() int {
	return len(s.t)
}

func (s sortT) Less(i, j int) bool {
	return s.t[i].Before(s.t[j])
}

func (s sortT) Swap(i, j int) {
	s.t[i], s.t[j] = s.t[j], s.t[i]
	s.b[i], s.b[j] = s.b[j], s.b[i]
}

func (e FinanceBrain) buildGraph() *bytes.Buffer {
	d := e.buildAccounts()
	n := time.Now()

	s := []chart.Series{}
	for account, bals := range d {
		x := []time.Time{}
		y := []float64{}
		for t, b := range bals {
			if n.Sub(t) > 365*24*time.Hour {
				continue
			}
			x = append(x, t)
			y = append(y, float64(b))
		}
		sort.Sort(sortT{x, y})
		ts := chart.TimeSeries{
			Name:    account,
			XValues: x,
			YValues: y,
		}
		s = append(s, ts)
	}

	graph := chart.Chart{
		Series: s,
		XAxis: chart.XAxis{
			Style: chart.Style{Show: true},
		},
		YAxis: chart.YAxis{
			Style: chart.Style{Show: true},
		},
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	b := &bytes.Buffer{}
	graph.Render(chart.PNG, b)
	return b
}

func (e FinanceBrain) handleReport(m Message) {

	total := e.totalBalance()

	b := e.buildGraph()
	// TODO clr
	// - Coverpage (Name,Title, Picture or something)
	// - Calculate savings (namely ratio of savings to not).
	// - What rent vs fun budget.
	// - Better graphs (aka can see all the lines).
	// - walk through my overview process basically
	// - holepunch

	bday, _ := time.Parse("2006/01/02", "2022/02/21")

	p := "<!doctype html><html><body>"
	p += "<header><h3>Colin's Financial Report</h3></header>"
	p += fmt.Sprintf("<img src=\"data:image/gif;base64,%s\" style=\"max-width: 80%%;\" />",
		base64.StdEncoding.EncodeToString(b.Bytes()))
	p += fmt.Sprintf("<p>Colin currently has %d$. On his 30th birthday Colin will be worth %d$ if he saves %d$/year, %d$ if %d$/year, %d$ if %d$/year given a 5%% maket return rate</p>", total,
		Interest(total, 100000, bday), 100000,
		Interest(total, 115000, bday), 115000,
		Interest(total, 130000, bday), 130000)

	p += "</body></html>"

	tf, err := ioutil.TempFile("", "report")
	if err != nil {
		log.Fatalf("Error opening temp file: %v", err)
	}

	if _, err = io.WriteString(tf, p); err != nil {
		log.Fatalf("Error writing html: %v", err)
	}
	out := Message{Destination: m.Source, Source: m.Destination}
	out.Body = fmt.Sprintf("Built a report")
	out.Attachments = []string{tf.Name()}
	e.W.SendMessage(out)
}
