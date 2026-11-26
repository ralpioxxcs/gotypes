package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSettingsRoundTrip(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "gotypes")
	if got := loadSettings(dir); got != (settings{}) {
		t.Fatalf("missing file should give zero settings, got %+v", got)
	}
	want := settings{Mode: "time 30", Theme: "nord", Tape: true, Lang: "korean"}
	if err := saveSettings(dir, want); err != nil {
		t.Fatal(err)
	}
	if got := loadSettings(dir); got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestBestPerModeAndLanguage(t *testing.T) {
	dir := t.TempDir()
	for _, r := range []record{
		{Mode: "words 30", Lang: "english", WPM: 80},
		{Mode: "words 30", Lang: "english", WPM: 95},
		{Mode: "words 30", Lang: "english", WPM: 70},
		{Mode: "words 30", Lang: "korean", WPM: 20, Taja: 300},
	} {
		if err := appendRecord(dir, r); err != nil {
			t.Fatal(err)
		}
	}
	f, _ := os.OpenFile(filepath.Join(dir, "history.jsonl"), os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString("{broken\n")
	f.Close()

	best := loadBest(dir)
	if best["words 30 english"] != 95 || best["words 30 korean"] != 300 || len(best) != 2 {
		t.Fatalf("best = %v", best)
	}
}
