package app

import (
	"fmt"
	"math"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/ralpioxxcs/gotypes/internal/engine"
)

type canvas struct {
	s  tcell.Screen
	th theme
}

func (c canvas) st(fg tcell.Color) tcell.Style {
	return tcell.StyleDefault.Background(c.th.bg).Foreground(fg)
}

func rw(r rune) int { return max(1, runewidth.RuneWidth(r)) }

func (c canvas) put(x, y int, r rune, st tcell.Style) int {
	c.s.SetContent(x, y, r, nil, st)
	return rw(r)
}

func (c canvas) print(x, y int, s string, st tcell.Style) int {
	for _, r := range s {
		x += c.put(x, y, r, st)
	}
	return x
}

func (c canvas) fill(x, y, w, h int, st tcell.Style) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			c.s.SetContent(i, j, ' ', nil, st)
		}
	}
}

func (c canvas) box(x, y, w, h int, st tcell.Style) {
	for i := x + 1; i < x+w-1; i++ {
		c.put(i, y, '─', st)
		c.put(i, y+h-1, '─', st)
	}
	for j := y + 1; j < y+h-1; j++ {
		c.put(x, j, '│', st)
		c.put(x+w-1, j, '│', st)
	}
	c.put(x, y, '╭', st)
	c.put(x+w-1, y, '╮', st)
	c.put(x, y+h-1, '╰', st)
	c.put(x+w-1, y+h-1, '╯', st)
}

// hints draws "key desc" pairs centered on row y.
func (c canvas) hints(w, y int, pairs ...string) {
	n := -3 // no gap after the last pair
	for i := 0; i+1 < len(pairs); i += 2 {
		n += runewidth.StringWidth(pairs[i]) + 1 + runewidth.StringWidth(pairs[i+1]) + 3
	}
	x := (w - n) / 2
	for i := 0; i+1 < len(pairs); i += 2 {
		x = c.print(x, y, pairs[i], c.st(c.th.text))
		x = c.print(x, y, " "+pairs[i+1]+"   ", c.st(c.th.sub))
	}
}

var bigFont = map[rune][5]string{
	'0': {"███", "█ █", "█ █", "█ █", "███"},
	'1': {" █ ", "██ ", " █ ", " █ ", "███"},
	'2': {"███", "  █", "███", "█  ", "███"},
	'3': {"███", "  █", "███", "  █", "███"},
	'4': {"█ █", "█ █", "███", "  █", "  █"},
	'5': {"███", "█  ", "███", "  █", "███"},
	'6': {"███", "█  ", "███", "█ █", "███"},
	'7': {"███", "  █", "  █", "  █", "  █"},
	'8': {"███", "█ █", "███", "█ █", "███"},
	'9': {"███", "█ █", "███", "  █", "███"},
	'%': {"█ █", "  █", " █ ", "█  ", "█ █"},
}

func (c canvas) big(x, y int, s string, st tcell.Style) {
	for _, r := range s {
		for row, line := range bigFont[r] {
			c.print(x, y+row, line, st)
		}
		x += 4
	}
}

type cell struct {
	r  rune
	st tcell.Style
}

// wordCells renders word wi plus its trailing space, with char states and caret.
func wordCells(e *engine.Engine, th theme, wi int) []cell {
	w := e.Words[wi]
	base := tcell.StyleDefault.Background(th.bg)
	wrongPast := wi < e.Cur && !w.Correct()
	n := max(len(w.Target), len(w.Typed))
	out := make([]cell, 0, n+1)
	for i := range n {
		var r rune
		if i < len(w.Target) {
			r = w.Target[i] // wrong chars show what should have been typed
		} else {
			r = w.Typed[i]
		}
		fg := th.sub
		switch w.State(i) {
		case engine.Correct:
			fg = th.text
		case engine.Incorrect:
			fg = th.err
		case engine.Extra:
			fg = th.errExtra
		}
		out = append(out, cell{r, base.Foreground(fg).Underline(wrongPast)})
	}
	out = append(out, cell{' ', base})
	if wi == e.Cur && !e.Finished {
		out[len(w.Typed)].st = base.Background(th.caret).Foreground(th.bg)
	}
	return out
}

func cellsWidth(cs []cell) int {
	n := 0
	for _, ce := range cs {
		n += rw(ce.r)
	}
	return n
}

// paragraph draws 3 wrapped lines, keeping the caret on the second line once
// past the first. Returns the caret position.
func (c canvas) paragraph(e *engine.Engine, x, y, width int) (cx, cy int) {
	type placed struct {
		wi int
		cs []cell
	}
	var lines [][]placed
	var line []placed
	used, curLine := 0, 0
	for wi := range e.Words {
		cs := wordCells(e, c.th, wi)
		wd := cellsWidth(cs)
		if used+wd-1 > width && len(line) > 0 {
			lines, line, used = append(lines, line), nil, 0
		}
		if wi == e.Cur {
			curLine = len(lines)
		}
		if wi > e.Cur && len(lines) > curLine+2 {
			break // enough lines below the caret
		}
		line, used = append(line, placed{wi, cs}), used+wd
	}
	lines = append(lines, line)
	first := max(0, curLine-1)
	for li := 0; li < 3 && first+li < len(lines); li++ {
		xx := x
		for _, p := range lines[first+li] {
			if p.wi == e.Cur {
				cx, cy = xx+cellsWidth(p.cs[:len(e.Words[e.Cur].Typed)]), y+li
			}
			for _, ce := range p.cs {
				xx += c.put(xx, y+li, ce.r, ce.st)
			}
		}
	}
	return cx, cy
}

// tape draws one line in [x, x+width): text starts left-aligned, the caret walks
// to the middle, then stays there while the text scrolls. Returns the caret position.
func (c canvas) tape(e *engine.Engine, x, y, width int) (int, int) {
	var stream []cell
	caret := 0
	for wi := max(0, e.Cur-15); wi < len(e.Words) && wi < e.Cur+25; wi++ {
		if wi == e.Cur {
			caret = len(stream) + len(e.Words[wi].Typed)
		}
		stream = append(stream, wordCells(e, c.th, wi)...)
	}
	cx := min(x+cellsWidth(stream[:caret]), x+width/2)
	half := float64(width) / 2
	xx := cx
	for i := caret; i < len(stream) && xx+rw(stream[i].r) <= x+width; i++ {
		ce := stream[i]
		fg, _, _ := ce.st.Decompose()
		if i > caret { // soft fade near the right edge only
			fg = blend(fg, c.th.bg, 0.8*(1-math.Min(1, float64(x+width-xx)/12)))
		}
		xx += c.put(xx, y, ce.r, ce.st.Foreground(fg))
	}
	xx = cx
	for i := caret - 1; i >= 0; i-- {
		ce := stream[i]
		if xx -= rw(ce.r); xx < x {
			break
		}
		fg, _, _ := ce.st.Decompose()
		c.put(xx, y, ce.r, ce.st.Foreground(blend(fg, c.th.bg, math.Min(1, 0.85*float64(cx-xx)/half))))
	}
	return cx, y
}

func (c canvas) modeBar(u *ui, x, y, width int) {
	xx, kind := x, ""
	for i, m := range modes {
		if m == quoteMode && !u.hasQuotes() {
			continue
		}
		val := c.st(c.th.sub)
		if i == u.mode {
			val = c.st(c.th.main).Bold(true)
		}
		if k := modeKind(m); k != kind {
			if kind != "" {
				xx += 2
			}
			kind = k
			if m == quoteMode { // no numbers: the label is the value
				xx = c.print(xx, y, k, val)
				continue
			}
			st := c.st(c.th.sub)
			if modeKind(modes[u.mode]) == k {
				st = c.st(c.th.text)
			}
			xx = c.print(xx, y, k+" ", st)
		}
		xx = c.print(xx, y, fmt.Sprint(m.N)+" ", val)
	}
	right := u.lang + " · " + c.th.name
	c.print(x+width-runewidth.StringWidth(right), y, right, c.st(c.th.sub))
}

func modeKind(m engine.Mode) string {
	switch {
	case m == quoteMode:
		return "quote"
	case m.Time:
		return "time"
	}
	return "words"
}

func progress(e *engine.Engine) string {
	if e.Mode.Time {
		return fmt.Sprint(max(0, e.Mode.N-int(e.Elapsed().Seconds())))
	}
	return fmt.Sprintf("%d/%d", e.Cur, e.Mode.N)
}

var blocks = []rune(" ▁▂▃▄▅▆▇█")

type chartGeom struct{ x, y, h, colW, start, n int }

// chart draws the per-second result: plot rows y..y+h-1, tooltip at y-1,
// error marks at y+h, x labels at y+h+1. Columns sweep in after the finish.
func (c canvas) chart(ps []engine.Second, sel int, since time.Duration, x, y, w, h int) chartGeom {
	th, n := c.th, len(ps)
	if n == 0 || h < 2 {
		return chartGeom{}
	}
	mx := 20.0
	for _, s := range ps {
		mx = math.Max(mx, math.Max(s.WPM, s.Raw))
	}
	mx = math.Ceil(mx/20) * 20

	const labelW = 4
	px, pw := x+labelW+1, w-labelW-1
	colW := min(max(pw/n, 1), 4)
	start := max(0, n-pw/colW)
	reveal := n
	if since < 700*time.Millisecond {
		reveal = int(float64(n)*float64(since)/float64(700*time.Millisecond)) + 1
	}
	if sel < 0 || sel >= n {
		sel = n - 1
	}

	sub := c.st(th.sub)
	c.print(x, y, fmt.Sprintf("%*d", labelW, int(mx)), sub)
	c.print(x, y+(h-1)/2, fmt.Sprintf("%*d", labelW, int(mx/2)), sub)
	c.print(x, y+h-1, fmt.Sprintf("%*d", labelW, 0), sub)
	barCol := blend(th.sub, th.bg, 0.45)
	bw := colW
	if colW >= 3 {
		bw = colW - 1
	}
	for i := start; i < min(n, reveal); i++ {
		cx := px + (i-start)*colW
		base := tcell.StyleDefault.Background(th.bg)
		if i == sel && reveal >= n {
			base = base.Background(th.subAlt)
			c.fill(cx, y, colW, h, base)
		}
		e8 := int(ps[i].Raw / mx * float64(h*8))
		for row := 0; row < h; row++ {
			if e := min(e8-row*8, 8); e > 0 {
				for k := range bw {
					c.put(cx+k, y+h-1-row, blocks[e], base.Foreground(barCol))
				}
			}
		}
		r := int(math.Round(ps[i].WPM / mx * float64(h-1)))
		c.put(cx+(bw-1)/2, y+h-1-r, '●', base.Foreground(th.main))
		if ps[i].Errors > 0 {
			c.put(cx+(bw-1)/2, y+h, '×', c.st(th.err))
		}
		if i == start || (i+1)%5 == 0 {
			c.print(cx, y+h+1, fmt.Sprint(i+1), sub)
		}
	}

	if reveal >= n {
		s := ps[sel]
		tx := c.print(px, y-1, fmt.Sprintf("%ds", sel+1), c.st(th.text).Bold(true))
		tx = c.print(tx+3, y-1, "wpm ", sub)
		tx = c.print(tx, y-1, fmt.Sprintf("%.0f", s.WPM), c.st(th.main))
		tx = c.print(tx+3, y-1, "raw ", sub)
		tx = c.print(tx, y-1, fmt.Sprintf("%.0f", s.Raw), c.st(th.text))
		tx = c.print(tx+3, y-1, "err ", sub)
		c.print(tx, y-1, fmt.Sprint(s.Errors), c.st(th.err))
	}
	lx := x + w - runewidth.StringWidth("● wpm  ▆ raw  × err")
	lx = c.print(lx, y-1, "●", c.st(th.main))
	lx = c.print(lx, y-1, " wpm  ", sub)
	lx = c.print(lx, y-1, "▆", c.st(barCol))
	lx = c.print(lx, y-1, " raw  ", sub)
	lx = c.print(lx, y-1, "×", c.st(th.err))
	c.print(lx, y-1, " err", sub)
	return chartGeom{px, y, h, colW, start, n}
}

func (u *ui) draw(s tcell.Screen, w, h int) {
	list, sel := u.menuList()
	th := themes[u.theme]
	if u.menuOpen && u.menuCat == catTheme && sel >= 0 {
		th = themes[themeIndex(list[sel])] // live preview
	}
	c := canvas{s, th}
	c.fill(0, 0, w, h, c.st(th.text))
	cw := min(w-4, 80)
	x0 := (w - cw) / 2
	e := u.e

	s.HideCursor()
	if !e.Finished {
		y0 := max(1, (h-8)/2)
		c.modeBar(u, x0, y0, cw)
		if e.Started {
			c.print(x0, y0+3, progress(e), c.st(th.main).Bold(true))
		}
		var cx, cy int
		if u.tape {
			cx, cy = c.tape(e, x0, y0+5, cw)
		} else {
			cx, cy = c.paragraph(e, x0, y0+5, cw)
		}
		if u.korean() && !u.menuOpen {
			s.ShowCursor(cx, cy) // the IME draws its in-progress syllable at the terminal cursor
		}
		c.hints(w, h-2, "tab", "restart", "esc", "menu", "ctrl+c", "quit")
	} else {
		y0 := max(1, (h-24)/2)
		c.modeBar(u, x0, y0, cw)
		if u.quote.Source != "" {
			c.print(x0, y0+1, "— "+u.quote.Source, c.st(th.sub))
		}
		label, speed, detail := "wpm", e.WPM(), ""
		if u.korean() {
			label, speed, detail = "타/분", e.Taja(), fmt.Sprintf("wpm %.0f   ", e.WPM())
		}
		c.print(x0, y0+3, label, c.st(th.sub))
		c.print(x0+18, y0+3, "acc", c.st(th.sub))
		c.big(x0, y0+4, fmt.Sprintf("%.0f", speed), c.st(th.main))
		c.big(x0+18, y0+4, fmt.Sprintf("%.0f%%", e.Accuracy()), c.st(th.main))
		ok, bad, extra := e.Counts()
		c.print(x0, y0+10, detail+fmt.Sprintf("raw %.0f   consistency %.0f%%   chars %d/%d/%d   time %.1fs",
			e.Raw(), e.Consistency(), ok, bad, extra, e.Elapsed().Seconds()), c.st(th.sub))
		switch {
		case u.saveErr != nil:
			c.print(x0, y0+11, "history not saved: "+u.saveErr.Error(), c.st(th.err))
		case u.pb && u.prevBest > 0:
			c.print(x0, y0+11, fmt.Sprintf("new personal best (was %.0f)", u.prevBest), c.st(th.main).Bold(true))
		case u.pb:
			c.print(x0, y0+11, "new personal best", c.st(th.main).Bold(true))
		default:
			c.print(x0, y0+11, fmt.Sprintf("personal best %.0f", u.prevBest), c.st(th.sub))
		}
		chH := min(8, h-5-(y0+13)) // room for error row, x labels, hints
		u.chart = c.chart(e.PerSecond(), u.sel, time.Since(u.finishedAt), x0, y0+13, cw, chH)
		c.hints(w, h-2, "←/→", "inspect", "tab", "restart", "esc", "menu", "ctrl+c", "quit")
	}

	if u.menuOpen {
		c.menu(u, list, sel, w, h)
	}
}

// menu: level 1 lists categories with their current value; level 2 lists one
// category's options under a search line.
func (c canvas) menu(u *ui, list []string, sel, w, h int) {
	th := c.th
	pane := c.st(th.text).Background(th.subAlt)
	dim := pane.Foreground(th.sub)
	hi := pane.Background(th.main).Foreground(th.bg)
	const bw = 34
	mx := (w - bw) / 2

	if u.menuCat < 0 {
		bh := len(categories) + 2
		my := (h - bh) / 2
		c.fill(mx, my, bw, bh, pane)
		c.box(mx, my, bw, bh, dim)
		c.print(mx+2, my, " settings ", dim)
		for i, cat := range categories {
			yy, st := my+1+i, pane
			if i == sel {
				st = hi
				c.fill(mx+1, yy, bw-2, 1, st)
			}
			c.print(mx+3, yy, cat.name, st)
			c.print(mx+14, yy, cat.current(u), st)
			c.put(mx+bw-3, yy, '›', st)
		}
		c.hints(w, my+bh+1, "↑↓", "move", "enter", "open", "esc", "close")
		return
	}

	cat := categories[u.menuCat]
	all := len(cat.options(u))
	vis := max(1, min(all, 10, h-8)) // box keeps its size while filtering
	bh := vis + 3
	my := (h - bh) / 2
	c.fill(mx, my, bw, bh, pane)
	c.box(mx, my, bw, bh, dim)
	c.print(mx+2, my, " "+cat.name+" ", dim)
	x := c.print(mx+3, my+1, "/ ", dim)
	x = c.print(x, my+1, u.query, pane)
	c.put(x, my+1, '▏', pane.Foreground(th.caret))
	if len(list) == 0 {
		c.print(mx+3, my+2, "no match", dim)
	}
	top := min(max(0, sel-vis/2), max(0, len(list)-vis))
	for i := top; i < min(len(list), top+vis); i++ {
		yy, st := my+2+i-top, pane
		if i == sel {
			st = hi
			c.fill(mx+1, yy, bw-2, 1, st)
		}
		mark := "  "
		if list[i] == cat.current(u) {
			mark = "● "
		}
		c.print(mx+3, yy, mark+list[i], st)
	}
	c.hints(w, my+bh+1, "↑↓", "move", "type", "search", "enter", "select", "esc", "back")
}
