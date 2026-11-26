package app

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

var logo = strings.Split(strings.TrimPrefix(`
 ▄▀▀█▀█▄  ▄▀▀███▄  ▄▀▀ ▀███▄  ▄▀▀  █▀██ █▀▀ ▀█▄   ▄▀▀█▀█▄   ▄▀▀ ▀ ██▄
■▄████░▒ ■▄████░▒ ■██▄████░▒ ■▄██▄ ■██▌ ■▄████░▒ ■▄████░▒   ■  ▄   ░▒
█▓██   ░ █▓██   ░   █▓██   ░ ▀▀████▄██▌ █▓██ ▄▄▀ █▓██▄▄▄    █  ▓ ▄▄ ░
█▒██ ▄▄▄ █▒██   █   █▒██        ▄ ▀▀██▌ █▒██     █▒██     █ ▀▀█▒  █
█░██▄▄▄█ █░██▄▄▄█   █░██      ▄▀█▄▄███  █░██     █░██▄▄▄█ ███  ░  █
▀██░███▀ ▀██░███▀   ███░     ███░███▀   ███░     ▀██░███▀ ▀██████░▀`, "\n"), "\n")

// logo draws the title art centered above content starting at row y0, fading
// from main to sub left to right; skipped when the screen has no room for it.
func (c canvas) logo(w, y0 int) {
	y := y0 - len(logo) - 1
	lw := 0
	for _, l := range logo {
		lw = max(lw, runewidth.StringWidth(l))
	}
	if y < 1 || lw > w {
		return
	}
	x := (w - lw) / 2
	for j, l := range logo {
		xx := x
		for _, r := range l {
			if r != ' ' {
				c.put(xx, y+j, r, c.st(blend(c.th.main, c.th.sub, 0.6*float64(xx-x)/float64(lw))))
			}
			xx += rw(r)
		}
	}
}
