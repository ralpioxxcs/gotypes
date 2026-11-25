package app

import (
	"slices"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestMenuSearchApplyAndBack(t *testing.T) {
	u := &ui{lists: map[string][]string{"english": {"a"}}, langs: []string{"english"}, lang: "english", mode: 1}
	u.restart()
	press := func(k tcell.Key, r rune) { u.key(tcell.NewEventKey(k, r, tcell.ModNone)) }

	press(tcell.KeyEscape, 0)
	for range catTheme {
		press(tcell.KeyDown, 0)
	}
	press(tcell.KeyEnter, 0)
	for _, r := range "TOK" { // case-insensitive
		press(tcell.KeyRune, r)
	}
	if list, sel := u.menuList(); !slices.Equal(list, []string{"tokyo night"}) || sel != 0 {
		t.Fatalf("filtered = %v sel %d", list, sel)
	}
	press(tcell.KeyEscape, 0) // back to categories, theme still selected
	if !u.menuOpen || u.menuCat != -1 || u.menuIdx != catTheme {
		t.Fatalf("esc should go back: open=%v cat=%d idx=%d", u.menuOpen, u.menuCat, u.menuIdx)
	}
	press(tcell.KeyEnter, 0)
	press(tcell.KeyRune, 'n')
	press(tcell.KeyRune, 'o')
	press(tcell.KeyEnter, 0)
	if u.menuOpen || themes[u.theme].name != "nord" {
		t.Fatalf("enter should apply nord and close: open=%v theme=%s", u.menuOpen, themes[u.theme].name)
	}

	if slices.Contains(categories[0].options(u), "quote") {
		t.Fatal("quote offered for a language without quotes")
	}
}
