package app

import (
	"fmt"
	"math"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/ralpioxxcs/gotypes/internal/engine"
	"github.com/ralpioxxcs/gotypes/internal/race"
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
		if !u.available(m) {
			continue
		}
		if k := m.kind(); k != kind {
			if kind != "" {
				xx += 2
			}
			kind = k
			st := c.st(c.th.sub)
			if modes[u.mode].kind() == k {
				st = c.st(c.th.text)
			}
			xx = c.print(xx, y, k+" ", st)
		}
		val := c.st(c.th.sub)
		if i == u.mode {
			val = c.st(c.th.main).Bold(true)
		}
		xx = c.print(xx, y, m.value()+" ", val)
	}
	c.print(x+width-runewidth.StringWidth(u.lang), y, u.lang, c.st(c.th.sub))
}

func progress(e *engine.Engine) string {
	if e.Mode.Time {
		return fmt.Sprint(max(0, e.Mode.N-int(e.Elapsed().Seconds())))
	}
	return fmt.Sprintf("%d/%d", e.Cur, e.Mode.N)
}

var eighths = []rune(" ▏▎▍▌▋▊▉")

// meter draws the live speed gauge: a bar filled to 1/8 cell, colored from sub
// to main along its length, a peak-hold tick, and the value on the right.
func (c canvas) meter(g gauge, unit string, x, y, w int) {
	label := fmt.Sprintf("%4.0f %s", g.shown, unit)
	bw := w - runewidth.StringWidth(label) - 2
	if bw < 4 || g.scale <= 0 {
		return
	}
	fill := g.shown / g.scale * float64(bw)
	peak := int(g.peak / g.scale * float64(bw))
	track := blend(c.th.sub, c.th.bg, 0.7)
	for i := range bw {
		col := blend(c.th.sub, c.th.main, float64(i)/float64(bw-1))
		switch f := fill - float64(i); {
		case f >= 1:
			c.put(x+i, y, '█', c.st(col))
		case f > 0 && int(f*8) > 0:
			c.put(x+i, y, eighths[int(f*8)], c.st(col))
		case i == peak && g.peak > 0:
			c.put(x+i, y, '┃', c.st(c.th.text))
		default:
			c.put(x+i, y, '─', c.st(track))
		}
	}
	c.print(x+bw+2, y, label, c.st(c.th.main).Bold(true))
}

// players draws the room as a titled box, one row per player; your row is
// highlighted. Before the start a row shows ready or not; during the race a
// progress bar and live speed, then place and time.
func (c canvas) players(snap race.Snapshot, unit string, x, y, w int) {
	th := c.th
	const nameW, labelW = 12, 10
	n := len(snap.Players)
	c.box(x, y, w, n+2, c.st(th.sub))
	title := fmt.Sprintf(" players %d/%d ", n, race.MaxPlayers)
	if snap.State == race.Running {
		title = " race "
	}
	c.print(x+2, y, title, c.st(th.text).Bold(true))

	ix, iw := x+2, w-4 // inside the border, one cell of padding
	bx, bw := ix+2+nameW+1, iw-2-nameW-1-labelW-1
	track := blend(th.sub, th.bg, 0.7)
	for i, p := range snap.Players {
		yy := y + 1 + i
		bg := th.bg
		if p.Me {
			bg = th.subAlt
			c.fill(x+1, yy, w-2, 1, c.st(th.text).Background(bg))
		}
		st := func(fg tcell.Color) tcell.Style { return c.st(fg).Background(bg) }
		name := st(th.text)
		if p.Me {
			name = st(th.main).Bold(true)
		}
		mark, markCol := '○', th.sub
		switch {
		case p.Left:
			mark, markCol = '×', th.err
		case p.Ready:
			mark, markCol = '●', th.main
		}
		c.put(ix, yy, mark, st(markCol))
		c.print(ix+2, yy, runewidth.Truncate(p.Name, nameW, "…"), name)

		label, labelSt := "not ready", st(th.sub)
		switch {
		case p.Left:
			label = "left"
		case p.Place > 0:
			label, labelSt = fmt.Sprintf("%s %.1fs", ordinal(p.Place), p.Time.Seconds()), st(th.main).Bold(true)
		case snap.State == race.Running:
			label, labelSt = fmt.Sprintf("%.0f %s", p.Speed, unit), st(th.text)
		case p.Ready:
			label, labelSt = "ready", st(th.main).Bold(true)
		}
		c.print(ix+iw-runewidth.StringWidth(label), yy, label, labelSt)

		if snap.State != race.Running || bw < 4 {
			continue
		}
		col := th.sub
		if p.Me {
			col = th.main
		}
		fill := p.Progress * float64(bw)
		for j := range bw {
			switch f := fill - float64(j); {
			case f >= 1:
				c.put(bx+j, yy, '█', st(col))
			case f > 0 && int(f*8) > 0:
				c.put(bx+j, yy, eighths[int(f*8)], st(col))
			default:
				c.put(bx+j, yy, '─', st(track))
			}
		}
	}
}

// raceStatus is the line above the text before typing starts.
func (c canvas) raceStatus(snap race.Snapshot, x, y int) {
	ready := snap.Me().Ready
	hi, dim := c.st(c.th.main).Bold(true), c.st(c.th.sub)
	switch snap.State {
	case race.Waiting:
		if ready {
			c.print(x, y, "waiting for another ready player…", dim)
		} else {
			c.print(x, y, "press enter when you're ready", hi)
		}
	case race.Counting:
		xx := c.print(x, y, fmt.Sprintf("starting in %d", int(math.Ceil(time.Until(snap.Start).Seconds()))), hi)
		if !ready {
			c.print(xx, y, "   press enter to join", dim)
		}
	default:
		c.print(x, y, "go!", hi)
	}
}

func ordinal(n int) string {
	if s := map[int]string{1: "st", 2: "nd", 3: "rd"}[n]; s != "" {
		return fmt.Sprint(n, s)
	}
	return fmt.Sprint(n, "th")
}

// chartGeom maps screen columns back to seconds for mouse hover.
type chartGeom struct{ x, y, h, dw, n int }

// second is the second drawn nearest to screen column sx.
func (g chartGeom) second(sx int) int {
	best := 0
	for i := range g.n {
		if abs(g.col(i)-(sx-g.x)) < abs(g.col(best)-(sx-g.x)) {
			best = i
		}
	}
	return best
}

// col is the plot cell (from the left edge) where second i is drawn.
func (g chartGeom) col(i int) int {
	if g.n < 2 {
		return 0
	}
	return i * (g.dw - 1) / (g.n - 1) / 2
}

func abs(v int) int { return max(v, -v) }

// brailleBit is the dot for (column 0-1, row 0-3) inside one braille cell (U+2800 + bits).
var brailleBit = [2][4]rune{{0x01, 0x02, 0x04, 0x40}, {0x08, 0x10, 0x20, 0x80}}

// chart draws the per-second result as braille line charts (2x4 dots per cell):
// wpm bright, raw dim. Plot rows y..y+h-1, tooltip at y-1, error marks at y+h,
// x labels at y+h+1. Lines sweep in left to right after the finish.
func (c canvas) chart(ps []engine.Second, sel int, since time.Duration, x, y, w, h int) chartGeom {
	th, n := c.th, len(ps)
	const labelW = 4
	px, pw := x+labelW+1, w-labelW-1
	if n == 0 || h < 2 || pw < 2 {
		return chartGeom{}
	}
	mx := 20.0
	for _, s := range ps {
		mx = math.Max(mx, math.Max(s.WPM, s.Raw))
	}
	mx = math.Ceil(mx/20) * 20

	dw, dh := pw*2, h*4
	xAt := func(i int) int {
		if n == 1 {
			return 0
		}
		return i * (dw - 1) / (n - 1)
	}
	yAt := func(v float64) int { return dh - 1 - int(math.Round(v/mx*float64(dh-1))) }
	reveal := dw
	if since < 700*time.Millisecond {
		reveal = int(float64(dw) * float64(since) / float64(700*time.Millisecond))
	}
	done := reveal >= dw
	if sel < 0 || sel >= n {
		sel = n - 1
	}

	// rasterize both series; a cell touched by the wpm line takes the wpm color
	bits, isWPM := make([]rune, pw*h), make([]bool, pw*h)
	plot := func(dx, dy int, wpm bool) {
		if dx < 0 || dx > reveal || dx >= dw || dy < 0 || dy >= dh {
			return
		}
		k := dy/4*pw + dx/2
		bits[k] |= brailleBit[dx%2][dy%4]
		isWPM[k] = isWPM[k] || wpm
	}
	line := func(x0, y0, x1, y1 int, wpm bool) { // Bresenham
		ax, ay, sx, sy := abs(x1-x0), -abs(y1-y0), 1, 1
		if x1 < x0 {
			sx = -1
		}
		if y1 < y0 {
			sy = -1
		}
		for e := ax + ay; ; {
			plot(x0, y0, wpm)
			if x0 == x1 && y0 == y1 {
				return
			}
			e2 := 2 * e
			if e2 >= ay {
				e, x0 = e+ay, x0+sx
			}
			if e2 <= ax {
				e, y0 = e+ax, y0+sy
			}
		}
	}
	for _, wpm := range []bool{false, true} {
		v := func(i int) float64 {
			if wpm {
				return ps[i].WPM
			}
			return ps[i].Raw
		}
		if n == 1 { // a lone point reads better as a level line
			line(0, yAt(v(0)), dw-1, yAt(v(0)), wpm)
		}
		for i := 1; i < n; i++ {
			line(xAt(i-1), yAt(v(i-1)), xAt(i), yAt(v(i)), wpm)
		}
	}

	sub := c.st(th.sub)
	rawCol, gridCol := blend(th.sub, th.bg, 0.35), blend(th.sub, th.bg, 0.75)
	selCol := xAt(sel) / 2
	for row := range h {
		for col := range pw {
			st := tcell.StyleDefault.Background(th.bg)
			if done && col == selCol {
				st = st.Background(th.subAlt)
			}
			k := row*pw + col
			switch {
			case bits[k] != 0:
				fg := rawCol
				if isWPM[k] {
					fg = th.main
				}
				c.put(px+col, y+row, 0x2800+bits[k], st.Foreground(fg))
			case row == 0 || row == (h-1)/2 || row == h-1:
				c.put(px+col, y+row, '┈', st.Foreground(gridCol))
			default:
				c.put(px+col, y+row, ' ', st)
			}
		}
	}

	c.print(x, y, fmt.Sprintf("%*d", labelW, int(mx)), sub)
	c.print(x, y+(h-1)/2, fmt.Sprintf("%*d", labelW, int(mx*(1-float64((h-1)/2*4+2)/float64(dh-1)))), sub)
	c.print(x, y+h-1, fmt.Sprintf("%*d", labelW, 0), sub)
	step := 5
	for float64(step)*float64(dw)/float64(max(n-1, 1))/2 < 4 {
		step *= 2 // keep x labels at least 4 cells apart
	}
	lastEnd := -1
	for i := range n {
		col := xAt(i) / 2
		if xAt(i) > reveal {
			break
		}
		if ps[i].Errors > 0 {
			c.put(px+col, y+h, '×', c.st(th.err))
		}
		if (i == 0 || (i+1)%step == 0) && col > lastEnd {
			lastEnd = c.print(px+col, y+h+1, fmt.Sprint(i+1), sub) - px
		}
	}

	if done {
		s := ps[sel]
		tx := c.print(px, y-1, fmt.Sprintf("%ds", sel+1), c.st(th.text).Bold(true))
		tx = c.print(tx+3, y-1, "wpm ", sub)
		tx = c.print(tx, y-1, fmt.Sprintf("%.0f", s.WPM), c.st(th.main))
		tx = c.print(tx+3, y-1, "raw ", sub)
		tx = c.print(tx, y-1, fmt.Sprintf("%.0f", s.Raw), c.st(th.text))
		tx = c.print(tx+3, y-1, "err ", sub)
		c.print(tx, y-1, fmt.Sprint(s.Errors), c.st(th.err))
	}
	lx := x + w - runewidth.StringWidth("━ wpm  ━ raw  × err")
	lx = c.print(lx, y-1, "━", c.st(th.main))
	lx = c.print(lx, y-1, " wpm  ", sub)
	lx = c.print(lx, y-1, "━", c.st(rawCol))
	lx = c.print(lx, y-1, " raw  ", sub)
	lx = c.print(lx, y-1, "×", c.st(th.err))
	c.print(lx, y-1, " err", sub)
	return chartGeom{px, y, h, dw, n}
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
	unit := "wpm"
	if u.korean() {
		unit = "타/분"
	}
	var snap race.Snapshot
	lanes := 0 // rows the race lanes take, with a gap above
	if u.seat != nil {
		snap = u.seat.Snapshot()
		lanes = len(snap.Players) + 3 // box borders and a gap above
	}
	if since := time.Since(u.openedAt); since < splashTime {
		s.HideCursor()
		c.splash(w, h, since, max(1, (h-8-lanes)/2)-len(logo)-1) // where the typing screen puts the logo
		return
	}

	s.HideCursor()
	if !e.Finished {
		y0 := max(1, (h-8-lanes)/2)
		c.logo(w, y0)
		c.modeBar(u, x0, y0, cw)
		switch {
		case e.Started:
			c.print(x0, y0+3, progress(e), c.st(th.main).Bold(true))
			c.meter(u.gauge, unit, x0+12, y0+3, cw-12)
		case u.seat != nil:
			c.raceStatus(snap, x0, y0+3)
		}
		if u.seat != nil {
			c.players(snap, unit, x0, y0+9, cw)
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
		switch {
		case u.held() && snap.Me().Ready:
			c.hints(w, h-2, "enter", "not ready", "tab", "new room", "esc", "menu", "ctrl+c", "quit")
		case u.held():
			c.hints(w, h-2, "enter", "ready", "tab", "new room", "esc", "menu", "ctrl+c", "quit")
		default:
			c.hints(w, h-2, "tab", "restart", "esc", "menu", "ctrl+c", "quit")
		}
	} else {
		y0 := max(1, (h-24-lanes)/2)
		c.logo(w, y0)
		c.modeBar(u, x0, y0, cw)
		if u.quote.Source != "" {
			c.print(x0, y0+1, "— "+u.quote.Source, c.st(th.sub))
		}
		speed, detail := e.WPM(), ""
		if u.korean() {
			speed, detail = e.Taja(), fmt.Sprintf("wpm %.0f   ", e.WPM())
		}
		c.print(x0, y0+3, unit, c.st(th.sub))
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
		chY := y0 + 13
		if lanes > 0 {
			c.players(snap, unit, x0, chY, cw)
			chY += lanes + 1 // box, gap and the chart's tooltip row
		}
		chH := min(8, h-5-chY) // room for error row, x labels, hints
		u.chart = c.chart(e.PerSecond(), u.sel, time.Since(u.finishedAt), x0, chY, cw, chH)
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
