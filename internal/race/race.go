// Package race matches players into rooms and shares their progress. It holds
// no goroutines: players report and read snapshots from their own loops.
package race

import (
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	Countdown  = 10 * time.Second // from the second ready player to the start
	MaxPlayers = 8
	nameLen    = 12
)

type State int

const (
	Waiting  State = iota // fewer than 2 ready players
	Counting              // start is set, not reached yet
	Running
)

type Player struct {
	Name     string
	Ready    bool          // only ready players race; the rest move to the next room at the start
	Progress float64       // 0..1
	Speed    float64       // wpm, or 타/분 in a Korean room
	Time     time.Duration // finish time since the start, 0 = still typing
	Place    int           // 1 = first to finish, 0 = not finished
	Left     bool          // disconnected mid-race; kept so the lanes stay put
	Me       bool          // set in snapshots for the reader's own lane
	base     string        // cleaned name before de-duplication, for moving rooms
}

type Snapshot struct {
	State   State
	Text    string
	Source  string
	Start   time.Time // zero while waiting
	Players []Player  // join order; once running, only the racers
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
// already racing. Callers hold h.mu.
func (h *Hub) seat(s *Seat, lang string) {
	r := h.open[lang]
	if r == nil || r.state(h.now()) == Running || len(r.players) >= MaxPlayers {
		text, source := h.pick(lang)
		r = &room{lang: lang, text: text, source: source}
		h.open[lang] = r
	}
	s.p.Name = r.unique(s.p.base)
	r.players = append(r.players, s.p)
	s.r = r
}

// settle moves a player who was not ready when the race started on to the
// next room. Callers hold h.mu; every Seat method settles first.
func (s *Seat) settle() {
	if s.r.state(s.h.now()) == Running && !s.p.Ready {
		s.r.remove(s.p)
		s.h.seat(s, s.r.lang)
	}
}

func (r *room) remove(p *Player) {
	for i, q := range r.players {
		if q == p {
			r.players = append(r.players[:i], r.players[i+1:]...)
			return
		}
	}
}

// recount starts the countdown at two ready players and stops it below that.
func (r *room) recount(now time.Time) {
	if r.state(now) == Running {
		return
	}
	n := 0
	for _, p := range r.players {
		if p.Ready {
			n++
		}
	}
	switch {
	case n < 2:
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
		if snap.State == Running && !p.Ready {
			continue // about to move on; see settle
		}
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
		s.p.Left = true
		return
	}
	s.r.remove(s.p)
	s.r.recount(s.h.now())
}
