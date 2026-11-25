// Package app is the terminal UI: one screen drawn by hand on top of
// internal/engine, which holds all typing logic.
package app

import (
	"encoding/json"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/ralpioxxcs/gotypes/data"
	"github.com/ralpioxxcs/gotypes/internal/engine"
	"github.com/rivo/tview"
)

// mode is a words/time test, or one quote of a length class; restart turns a
// quote into "N words" for the engine.
type mode struct {
	engine.Mode
	quote string // "short" | "medium" | "long", empty for words/time
}

var modes = []mode{
	{Mode: engine.Mode{N: 15}}, {Mode: engine.Mode{N: 30}}, {Mode: engine.Mode{N: 60}}, {Mode: engine.Mode{N: 120}},
	{Mode: engine.Mode{Time: true, N: 15}}, {Mode: engine.Mode{Time: true, N: 30}}, {Mode: engine.Mode{Time: true, N: 60}},
	{quote: "short"}, {quote: "medium"}, {quote: "long"},
}

func (m mode) kind() string {
	switch {
	case m.quote != "":
		return "quote"
	case m.Time:
		return "time"
	}
	return "words"
}

func (m mode) value() string {
	if m.quote != "" {
		return m.quote
	}
	return strconv.Itoa(m.N)
}

func (m mode) name() string { return m.kind() + " " + m.value() }

type quote struct {
	Text   string `json:"text"`
	Source string `json:"source"`
}

// quoteLen classes a quote by characters: short ≤ 60, medium ≤ 250, long beyond.
func quoteLen(q quote) string {
	switch n := utf8.RuneCountInString(q.Text); {
	case n <= 60:
		return "short"
	case n <= 250:
		return "medium"
	}
	return "long"
}

func (u *ui) quotesFor(length string) []quote {
	var out []quote
	for _, q := range u.quotes[u.lang] {
		if quoteLen(q) == length {
			out = append(out, q)
		}
	}
	return out
}

// available hides quote lengths the current language has no quotes for.
func (u *ui) available(m mode) bool { return m.quote == "" || len(u.quotesFor(m.quote)) > 0 }

// category is one entry of the settings menu. Values are labels, so
// filtering and saving work on the same strings the user sees.
type category struct {
	name    string
	options func(u *ui) []string
	current func(u *ui) string
	apply   func(u *ui, v string)
}

const catTheme = 3 // index in categories; the one that previews on hover

var categories = []category{
	{"mode",
		func(u *ui) []string {
			var out []string
			for _, m := range modes {
				if u.available(m) {
					out = append(out, m.name())
				}
			}
			return out
		},
		func(u *ui) string { return modes[u.mode].name() },
		func(u *ui, v string) { u.mode = modeIndex(v); u.restart() }},
	{"text",
		func(*ui) []string { return []string{"paragraph", "tape"} },
		func(u *ui) string {
			if u.tape {
				return "tape"
			}
			return "paragraph"
		},
		func(u *ui, v string) { u.tape = v == "tape" }},
	{"language",
		func(u *ui) []string { return u.langs },
		func(u *ui) string { return u.lang },
		func(u *ui, v string) { u.lang = v; u.restart() }},
	{"theme",
		func(*ui) []string {
			out := make([]string, len(themes))
			for i, t := range themes {
				out[i] = t.name
			}
			return out
		},
		func(u *ui) string { return themes[u.theme].name },
		func(u *ui, v string) { u.theme = themeIndex(v) }},
}

// modeIndex finds a mode by name; unknown names give words 30.
func modeIndex(name string) int {
	if i := slices.IndexFunc(modes, func(m mode) bool { return m.name() == name }); i >= 0 {
		return i
	}
	return 1
}

func themeIndex(name string) int {
	return max(0, slices.IndexFunc(themes, func(t theme) bool { return t.name == name }))
}

type ui struct {
	app        *tview.Application
	lists      map[string][]string // words per language
	langs      []string            // sorted keys of lists
	quotes     map[string][]quote  // per language; may be missing
	quote      quote               // the quote being typed, if any
	e          *engine.Engine
	mode       int // index into modes
	theme      int
	tape       bool
	lang       string
	sel        int // selected second on the result chart, -1 = last
	finishedAt time.Time
	chart      chartGeom // last drawn chart, for mouse hit-testing

	menuOpen bool
	menuCat  int    // -1 = category list, else index into categories
	menuIdx  int    // selection within the current (filtered) list
	query    string // search text inside a category

	dir      string             // where settings/history live; "" = saving off
	best     map[string]float64 // best score per mode+language
	prevBest float64            // best before the shown result
	pb       bool               // shown result is a new personal best
	saveErr  error
}

func (u *ui) restart() {
	if !u.available(modes[u.mode]) {
		u.mode = 1 // words 30
	}
	m := modes[u.mode]
	words := u.lists[u.lang]
	next := func() string { return words[rand.IntN(len(words))] }
	u.quote = quote{}
	if m.quote != "" {
		qs := u.quotesFor(m.quote)
		u.quote = qs[rand.IntN(len(qs))]
		ws := strings.Fields(u.quote.Text)
		m.N = len(ws)
		next = func() string { w := ws[0]; ws = ws[1:]; return w }
	}
	u.e = engine.New(m.Mode, next, time.Now)
	u.sel, u.finishedAt, u.pb, u.saveErr = -1, time.Time{}, false, nil
}

func (u *ui) korean() bool { return u.lang == "korean" }

// recordFinish runs once per finished test: PB check and history append.
func (u *ui) recordFinish() {
	if !u.e.Finished || !u.finishedAt.IsZero() {
		return
	}
	u.finishedAt = time.Now()
	e := u.e
	r := record{Time: u.finishedAt, Mode: modes[u.mode].name(), Lang: u.lang, WPM: e.WPM(), Raw: e.Raw(),
		Acc: e.Accuracy(), Consistency: e.Consistency(), Seconds: e.Elapsed().Seconds()}
	if u.korean() {
		r.Taja = e.Taja()
	}
	u.prevBest = u.best[r.key()]
	if u.pb = r.score() > u.prevBest; u.pb {
		u.best[r.key()] = r.score()
	}
	if u.dir != "" {
		u.saveErr = appendRecord(u.dir, r)
	}
}

func (u *ui) saveSettings() {
	if u.dir != "" {
		// shortcut: a failed settings write is dropped silently; only defaults are lost.
		_ = saveSettings(u.dir, settings{Mode: modes[u.mode].name(), Theme: themes[u.theme].name, Tape: u.tape, Lang: u.lang})
	}
}

// menuList is what the menu currently shows, and the clamped selection (-1 if empty).
func (u *ui) menuList() ([]string, int) {
	var list []string
	if u.menuCat < 0 {
		for _, c := range categories {
			list = append(list, c.name)
		}
	} else {
		q := strings.ToLower(u.query)
		for _, o := range categories[u.menuCat].options(u) {
			if strings.Contains(strings.ToLower(o), q) {
				list = append(list, o)
			}
		}
	}
	return list, min(u.menuIdx, len(list)-1)
}

func (u *ui) openCategory(i int) {
	u.menuCat, u.query = i, ""
	u.menuIdx = max(0, slices.Index(categories[i].options(u), categories[i].current(u)))
}

func (u *ui) menuKey(ev *tcell.EventKey) {
	list, sel := u.menuList()
	switch ev.Key() {
	case tcell.KeyEscape, tcell.KeyLeft:
		if u.menuCat < 0 {
			u.menuOpen = false
		} else {
			u.menuIdx, u.menuCat = u.menuCat, -1
		}
	case tcell.KeyUp, tcell.KeyDown:
		if len(list) > 0 {
			d := 1
			if ev.Key() == tcell.KeyUp {
				d = -1
			}
			u.menuIdx = (sel + d + len(list)) % len(list)
		}
	case tcell.KeyEnter, tcell.KeyRight:
		switch {
		case sel < 0:
		case u.menuCat < 0:
			u.openCategory(sel)
		default:
			categories[u.menuCat].apply(u, list[sel])
			u.menuOpen = false
			u.saveSettings()
		}
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if r := []rune(u.query); len(r) > 0 {
			u.query, u.menuIdx = string(r[:len(r)-1]), 0
		}
	case tcell.KeyRune:
		if u.menuCat >= 0 {
			u.query, u.menuIdx = u.query+string(ev.Rune()), 0
		}
	}
}

func (u *ui) moveSel(d int) {
	if !u.e.Finished || u.chart.n == 0 {
		return
	}
	if u.sel < 0 {
		u.sel = u.chart.n - 1
	}
	u.sel = min(max(u.sel+d, 0), u.chart.n-1)
}

func (u *ui) key(ev *tcell.EventKey) {
	if ev.Key() == tcell.KeyCtrlC {
		u.app.Stop()
		return
	}
	if u.menuOpen {
		u.menuKey(ev)
		return
	}
	switch ev.Key() {
	case tcell.KeyTab:
		u.restart()
	case tcell.KeyEscape:
		u.menuOpen, u.menuCat, u.menuIdx = true, -1, 0
	case tcell.KeyLeft:
		u.moveSel(-1)
	case tcell.KeyRight:
		u.moveSel(1)
	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if ev.Modifiers()&(tcell.ModAlt|tcell.ModCtrl) != 0 {
			u.e.DeleteWord()
		} else {
			u.e.Backspace()
		}
	case tcell.KeyCtrlW:
		u.e.DeleteWord()
	case tcell.KeyRune:
		u.e.Type(ev.Rune())
	}
	u.recordFinish()
}

func (u *ui) mouse(ev *tcell.EventMouse, a tview.MouseAction) (*tcell.EventMouse, tview.MouseAction) {
	g := u.chart
	if a == tview.MouseMove && u.e.Finished && !u.menuOpen && g.n > 0 {
		x, y := ev.Position()
		if y >= g.y-1 && y <= g.y+g.h+1 && x >= g.x {
			if i := g.start + (x-g.x)/g.colW; i < g.n {
				u.sel = i
			}
		}
	}
	return nil, 0
}

// Run starts the app and blocks until it quits.
func Run() {
	u := &ui{app: tview.NewApplication(), dir: storeDir()}
	// embedded at build time, so only a broken build panics here
	if err := json.Unmarshal(data.Words, &u.lists); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(data.Quotes, &u.quotes); err != nil {
		panic(err)
	}
	for l := range u.lists {
		u.langs = append(u.langs, l)
	}
	slices.Sort(u.langs)

	s := loadSettings(u.dir)
	u.mode = modeIndex(s.Mode)
	u.theme = themeIndex(s.Theme)
	u.tape, u.lang = s.Tape, s.Lang
	if _, ok := u.lists[u.lang]; !ok {
		u.lang = "english"
	}
	u.best = loadBest(u.dir)
	u.restart()

	root := tview.NewBox().SetDrawFunc(func(s tcell.Screen, x, y, w, h int) (int, int, int, int) {
		u.draw(s, w, h)
		return x, y, w, h
	})
	u.app.SetInputCapture(func(ev *tcell.EventKey) *tcell.EventKey { u.key(ev); return nil })
	u.app.SetMouseCapture(u.mouse)
	u.app.EnableMouse(true)
	go func() {
		// drives the clock, time-mode end and the result animation; all state stays on the UI goroutine
		for range time.NewTicker(100 * time.Millisecond).C {
			u.app.QueueUpdateDraw(func() { u.e.Tick(); u.recordFinish() })
		}
	}()
	if err := u.app.SetRoot(root, true).Run(); err != nil {
		panic(err)
	}
}
