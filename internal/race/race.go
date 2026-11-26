// Package race matches players into rooms and shares their progress. It holds
// no goroutines: players report and read snapshots from their own loops.
package race

import (
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/ralpioxxcs/gotypes/internal/engine"
)

const (
	Countdown  = 10 * time.Second // from everyone being ready to the start
	MaxPlayers = 8
	nameLen    = 12
)

type State int

const (
	Waiting  State = iota // someone is not ready, or fewer than 2 players
	Counting              // start is set, not reached yet
	Running
)

type Player struct {
	Name     string
	Ready    bool          // the countdown runs while everyone is ready
	Progress float64       // 0..1
	Speed    float64       // wpm, or 타/분 in a Korean room
	Time     time.Duration // finish time since the start, 0 = still typing
	Place    int           // 1 = first to finish, 0 = not finished
	Left     bool          // disconnected mid-race; kept so the lanes stay put
	Me       bool          // set in snapshots for the reader's own lane
	base     string        // cleaned name before de-duplication, for moving rooms
	cpm      float64       // a bot's characters per minute; 0 for people
}

// botWPM are the speeds of bots added one after another in a room.
var botWPM = []int{40, 60, 80, 100}

type Snapshot struct {
	State   State
	Text    string
	Source  string
	Start   time.Time // zero while waiting
	Players []Player  // join order
}

// Me is the reader's own entry.
func (s Snapshot) Me() Player {
	for _, p := range s.Players {
		if p.Me {
			return p
		}
	}
	return Player{}
}

type room struct {
	lang, text, source string
	start              time.Time
	players            []*Player
	finished           int
	next               *room // where its players go for a rematch
}

type Hub struct {
	mu   sync.Mutex
	now  func() time.Time
	pick func(lang string) (text, source string)
	open map[string]*room // per language: the room new players join
}

// NewHub matches players; pick chooses a room's text, now is the clock.
func NewHub(pick func(lang string) (text, source string), now func() time.Time) *Hub {
	return &Hub{now: now, pick: pick, open: map[string]*room{}}
}

// Seat is one player's place in a room.
type Seat struct {
	h *Hub
	r *room
	p *Player
}

func (r *room) state(now time.Time) State {
	switch {
	case r.start.IsZero():
		return Waiting
	case now.Before(r.start):
		return Counting
	}
	return Running
}

// Join puts a player, not ready yet, in the open room for lang.
func (h *Hub) Join(name, lang string) *Seat {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := &Seat{h: h, p: &Player{base: clean(name)}}
	h.seat(s, lang)
	return s
}

// seat places s in the open room for lang, opening one if that is full or
// already counting down. Callers hold h.mu.
func (h *Hub) seat(s *Seat, lang string) {
	r := h.open[lang]
	if r == nil || r.state(h.now()) != Waiting || len(r.players) >= MaxPlayers {
		text, source := h.pick(lang)
		r = &room{lang: lang, text: text, source: source}
		h.open[lang] = r
	}
	s.p.Name = r.unique(s.p.base)
	r.players = append(r.players, s.p)
	s.r = r
}

// settle brings the room up to now. Callers hold h.mu; every Seat method
// settles first.
func (s *Seat) settle() { s.r.advance(s.h.now()) }

func (r *room) remove(p *Player) {
	for i, q := range r.players {
		if q == p {
			r.players = append(r.players[:i], r.players[i+1:]...)
			return
		}
	}
}

// recount starts the countdown once all of at least two players (bots count,
// but one must be a person) are ready, and stops it when that no longer holds.
func (r *room) recount(now time.Time) {
	if r.state(now) == Running {
		return
	}
	n, people := 0, 0
	for _, p := range r.players {
		if p.Ready {
			n++
			if p.cpm == 0 {
				people++
			}
		}
	}
	switch {
	case n < 2 || n < len(r.players) || people == 0:
		r.start = time.Time{}
	case r.start.IsZero():
		r.start = now.Add(Countdown)
	}
}

// SetReady marks the player ready or not; it has no effect once racing.
func (s *Seat) SetReady(ready bool) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	s.settle()
	if s.r.state(s.h.now()) != Running {
		s.p.Ready = ready
		s.r.recount(s.h.now())
	}
}

// clean keeps a name to letters, digits, '-', '_' and '.', at most nameLen runes.
func clean(name string) string {
	var b []rune
	for _, c := range name {
		if len(b) == nameLen {
			break
		}
		if unicode.IsLetter(c) || unicode.IsDigit(c) || strings.ContainsRune("-_.", c) {
			b = append(b, c)
		}
	}
	if len(b) == 0 {
		return "player"
	}
	return string(b)
}

func (r *room) unique(name string) string {
	taken := func(n string) bool {
		for _, p := range r.players {
			if p.Name == n {
				return true
			}
		}
		return false
	}
	out := name
	for i := 2; taken(out); i++ {
		out = name + "#" + strconv.Itoa(i)
	}
	return out
}

// Report updates the player's lane. Reports before the start are ignored, and
// the first finished report fixes the time and place.
func (s *Seat) Report(progress, speed float64, finished bool) {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	s.settle()
	now := s.h.now()
	if s.r.state(now) != Running || s.p.Place > 0 || s.p.Left {
		return
	}
	s.p.Progress, s.p.Speed = min(max(progress, 0), 1), speed
	if finished {
		s.r.finished++
		s.p.Progress, s.p.Place, s.p.Time = 1, s.r.finished, max(now.Sub(s.r.start), 1)
	}
}

func (s *Seat) Snapshot() Snapshot {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	s.settle()
	r := s.r
	snap := Snapshot{State: r.state(s.h.now()), Text: r.text, Source: r.source, Start: r.start}
	for _, p := range r.players {
		cp := *p
		cp.Me = p == s.p
		snap.Players = append(snap.Players, cp)
	}
	return snap
}

// Leave frees the seat: a racer's lane stays marked as left, anyone else is
// removed (which may stop the countdown).
func (s *Seat) Leave() {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	s.settle()
	if s.r.state(s.h.now()) == Running {
		s.p.Left = s.p.Place == 0 // a finished lane keeps its place
		return
	}
	s.r.remove(s.p)
	if !slices.ContainsFunc(s.r.players, func(p *Player) bool { return p.cpm == 0 }) {
		s.r.players = nil // nobody left to race the bots
	}
	s.r.recount(s.h.now())
}

// AddBot puts the next bot (see botWPM) in the room before the race starts.
func (s *Seat) AddBot() {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	s.settle()
	r := s.r
	if r.state(s.h.now()) == Running || len(r.players) >= MaxPlayers {
		return
	}
	bots := 0
	for _, p := range r.players {
		if p.cpm > 0 {
			bots++
		}
	}
	wpm := botWPM[bots%len(botWPM)]
	r.addBot(&Player{base: "bot" + strconv.Itoa(wpm), cpm: float64(wpm * 5)})
	r.recount(s.h.now())
}

// RemoveBot takes the last added bot out of the room before the race starts.
func (s *Seat) RemoveBot() {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	s.settle()
	r := s.r
	if r.state(s.h.now()) == Running {
		return
	}
	for i := len(r.players) - 1; i >= 0; i-- {
		if r.players[i].cpm > 0 {
			r.remove(r.players[i])
			break
		}
	}
	r.recount(s.h.now())
}

func (r *room) addBot(b *Player) {
	units, runes := 0, 0
	for _, c := range r.text {
		units, runes = units+engine.Strokes(c), runes+1
	}
	b.Name, b.Ready, b.Speed = r.unique(b.base), true, b.cpm/5
	if r.lang == "korean" && runes > 0 { // 타/분: keystrokes, not characters
		b.Speed = b.cpm * float64(units) / float64(runes)
	}
	r.players = append(r.players, b)
}

// advance moves the bots to where their speed has taken them by now. Bots
// that got to the end are placed in the order they did, before any person
// reporting at now, so places stay true however rarely this runs.
func (r *room) advance(now time.Time) {
	if r.state(now) != Running {
		return
	}
	total := float64(utf8.RuneCountInString(r.text))
	var done []*Player
	for _, p := range r.players {
		if p.cpm == 0 || p.Place > 0 || total == 0 {
			continue
		}
		p.Progress = min(now.Sub(r.start).Minutes()*p.cpm/total, 1)
		if p.Progress == 1 {
			p.Time = time.Duration(total / p.cpm * float64(time.Minute))
			done = append(done, p)
		}
	}
	slices.SortFunc(done, func(a, b *Player) int { return int(a.Time - b.Time) })
	for _, p := range done {
		r.finished++
		p.Place = r.finished
	}
}

// Rematch moves a player who has finished to the room's next race, together
// with whoever else from the room does the same (and its bots), ready to go.
// It reports whether the player moved.
func (s *Seat) Rematch() bool {
	s.h.mu.Lock()
	defer s.h.mu.Unlock()
	s.settle()
	now, r := s.h.now(), s.r
	if r.state(now) != Running || s.p.Place == 0 {
		return false
	}
	if n := r.next; n == nil || n.state(now) == Running || len(n.players) >= MaxPlayers {
		text, source := s.h.pick(r.lang)
		r.next = &room{lang: r.lang, text: text, source: source}
		for _, p := range r.players {
			if p.cpm > 0 {
				r.next.addBot(&Player{base: p.base, cpm: p.cpm})
			}
		}
	}
	s.p = &Player{base: s.p.base, Ready: true}
	s.p.Name = r.next.unique(s.p.base)
	r.next.players = append(r.next.players, s.p)
	s.r = r.next
	s.r.recount(now)
	return true
}
