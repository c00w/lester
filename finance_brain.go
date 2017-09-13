package lester

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"

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
}

func badInterest(initial, deposit, years int64) int64 {
	i := float64(initial)
	for d := int64(0); d < years; d++ {
		i *= 1.05
		i += float64(deposit)
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

func (e FinanceBrain) handleReport(m Message) {

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

	i, err := png.Decode(b)
	if err != nil {
		log.Fatalf("unable to open png file: %v", err)
	}

	oi := image.NewRGBA(image.Rect(
		i.Bounds().Min.X,
		i.Bounds().Min.Y,
		i.Bounds().Max.X,
		i.Bounds().Max.Y+200,
	))
	draw.Draw(oi, image.Rect(
		oi.Bounds().Min.X,
		oi.Bounds().Min.Y,
		oi.Bounds().Max.X,
		oi.Bounds().Max.Y,
	), image.NewUniform(color.RGBA{255, 255, 255, 255}), image.Point{}, draw.Src)
	draw.Draw(oi, image.Rect(
		i.Bounds().Min.X,
		i.Bounds().Min.Y+200,
		i.Bounds().Max.X,
		i.Bounds().Max.Y+200,
	), i, image.Point{}, draw.Src)

	col := color.RGBA{0, 0, 0, 255}
	point := fixed.P(20, 50)
	dr := &font.Drawer{
		Dst:  oi,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}
	dr.DrawString("Colin's Financial Report")
	dr.Dot = fixed.P(20, 70)
	dr.DrawString(fmt.Sprintf("Produced on %s", time.Now().Format("Mon Jan 2 2006")))

	tf, err := ioutil.TempFile("", "report")
	if err != nil {
		log.Fatalf("Error opening temp file: %v", err)
	}

	png.Encode(tf, oi)
	out := Message{Destination: m.Source, Source: m.Destination}
	out.Body = fmt.Sprintf("Built a report")
	out.Attachments = []string{tf.Name()}
	e.W.SendMessage(out)
}
