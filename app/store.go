package app

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/ralpioxxcs/gotypes/internal/engine"
)

// settings survive restarts. Stored by value/name so reordering lists can't misapply them.
type settings struct {
	Mode  engine.Mode `json:"mode"`
	Theme string      `json:"theme"`
	Tape  bool        `json:"tape"`
	Lang  string      `json:"lang"`
}

// record is one finished test, one JSON line in history.jsonl.
type record struct {
	Time        time.Time `json:"time"`
	Mode        string    `json:"mode"`
	Lang        string    `json:"lang"`
	WPM         float64   `json:"wpm"`
	Taja        float64   `json:"taja,omitempty"`
	Raw         float64   `json:"raw"`
	Acc         float64   `json:"acc"`
	Consistency float64   `json:"consistency"`
	Seconds     float64   `json:"seconds"`
}

func (r record) key() string { return r.Mode + " " + r.Lang }

// score ranks results: 타/분 for Korean, wpm otherwise.
func (r record) score() float64 {
	if r.Lang == "korean" {
		return r.Taja
	}
	return r.WPM
}

// storeDir is "" when the OS has no config dir (e.g. no $HOME in a container); saving is then off.
func storeDir() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "gotypes")
}

func loadSettings(dir string) (s settings) {
	if b, err := os.ReadFile(filepath.Join(dir, "config.json")); err == nil {
		_ = json.Unmarshal(b, &s) // missing or broken file → defaults
	}
	return s
}

func saveSettings(dir string, s settings) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, _ := json.Marshal(s) // plain struct, cannot fail
	return os.WriteFile(filepath.Join(dir, "config.json"), b, 0o644)
}

// loadBest returns the best score per mode+language; unreadable lines are skipped.
func loadBest(dir string) map[string]float64 {
	best := map[string]float64{}
	f, err := os.Open(filepath.Join(dir, "history.jsonl"))
	if err != nil {
		return best
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var r record
		if json.Unmarshal(sc.Bytes(), &r) == nil {
			best[r.key()] = max(best[r.key()], r.score())
		}
	}
	return best
}

func appendRecord(dir string, r record) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(dir, "history.jsonl"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	b, _ := json.Marshal(r)
	_, err = f.Write(append(b, '\n'))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err
}
