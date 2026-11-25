// Package app is the terminal UI: one screen drawn by hand on top of
// internal/engine, which holds all typing logic.
package app

import (
	"encoding/json"
	"math/rand/v2"
	"slices"
	"strconv"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/ralpioxxcs/gotypes/data"
	"github.com/ralpioxxcs/gotypes/internal/engine"
	"github.com/rivo/tview"
)

var modes = []engine.Mode{
	{N: 15}, {N: 30}, {N: 60}, {N: 120},
	{Time: true, N: 15}, {Time: true, N: 30}, {Time: true, N: 60},
}

type itemKind int

const (
	itemHeader itemKind = iota
	itemMode
	itemText
	itemLang
	itemTheme
)

type menuItem struct {
	label string
	kind  itemKind
	idx   int
}

func (u *ui) menuItems() []menuItem {
	items := []menuItem{{label: "mode"}}
	for i, m := range modes {
		items = append(items, menuItem{modeName(m), itemMode, i})
	}
	items = append(items, menuItem{label: "text"},
		menuItem{"paragraph", itemText, 0}, menuItem{"tape", itemText, 1},
		menuItem{label: "language"})
	for i, l := range u.langs {
		items = append(items, menuItem{l, itemLang, i})
	}
	items = append(items, menuItem{label: "theme"})
	for i, t := range themes {
		items = append(items, menuItem{t.name, itemTheme, i})
	}
	return items
}

func modeName(m engine.Mode) string { return modeKind(m) + " " + strconv.Itoa(m.N) }

type ui struct {
	app        *tview.Application
	lists      map[string][]string // words per language
	langs      []string            // sorted keys of lists
	e          *engine.Engine
	mode       int // index into modes
	theme      int
	tape       bool
	lang       string
	menuOpen   bool
	menuIdx    int
	sel        int // selected second on the result chart, -1 = last
	finishedAt time.Time
	chart      chartGeom // last drawn chart, for mouse hit-testing

	dir      string             // where settings/history live; "" = saving off
	best     map[string]float64 // best score per mode+language
	prevBest float64            // best before the shown result
	pb       bool               // shown result is a new personal best
	saveErr  error
}

func (u *ui) restart() {
	words := u.lists[u.lang]
	u.e = engine.New(modes[u.mode], func() string { return words[rand.IntN(len(words))] }, time.Now)
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
	r := record{Time: u.finishedAt, Mode: modeName(e.Mode), Lang: u.lang, WPM: e.WPM(), Raw: e.Raw(),
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
		_ = saveSettings(u.dir, settings{Mode: modes[u.mode], Theme: themes[u.theme].name, Tape: u.tape, Lang: u.lang})
	}
}

func (u *ui) active(it menuItem) bool {
	switch it.kind {
	case itemMode:
		return it.idx == u.mode
	case itemText:
		return (it.idx == 1) == u.tape
	case itemLang:
		return u.langs[it.idx] == u.lang
	case itemTheme:
		return it.idx == u.theme
	}
	return false
}

func (u *ui) moveMenu(d int) {
	items := u.menuItems()
	for {
		u.menuIdx = (u.menuIdx + d + len(items)) % len(items)
		if items[u.menuIdx].kind != itemHeader {
			return
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
		switch ev.Key() {
		case tcell.KeyEscape:
			u.menuOpen = false
		case tcell.KeyUp:
			u.moveMenu(-1)
		case tcell.KeyDown:
			u.moveMenu(1)
		case tcell.KeyEnter:
			switch it := u.menuItems()[u.menuIdx]; it.kind {
			case itemMode:
				u.mode = it.idx
				u.restart()
			case itemText:
				u.tape = it.idx == 1
			case itemLang:
				u.lang = u.langs[it.idx]
				u.restart()
			case itemTheme:
				u.theme = it.idx
			}
			u.menuOpen = false
			u.saveSettings()
		}
		return
	}
	switch ev.Key() {
	case tcell.KeyTab:
		u.restart()
	case tcell.KeyEscape:
		u.menuOpen, u.menuIdx = true, 1+u.mode
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
	if err := json.Unmarshal(data.Words, &u.lists); err != nil {
		panic(err) // embedded at build time, so only a broken build gets here
	}
	for l := range u.lists {
		u.langs = append(u.langs, l)
	}
	slices.Sort(u.langs)

	s := loadSettings(u.dir)
	if u.mode = slices.Index(modes, s.Mode); u.mode < 0 {
		u.mode = 1 // words 30
	}
	u.theme = max(0, slices.IndexFunc(themes, func(t theme) bool { return t.name == s.Theme }))
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
