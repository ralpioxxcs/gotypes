package app

import "github.com/gdamore/tcell/v2"

// theme uses monkeytype's color roles.
type theme struct {
	name                                              string
	bg, main, caret, sub, subAlt, text, err, errExtra tcell.Color
}

func hex(s string) tcell.Color { return tcell.GetColor(s) } // needs #rrggbb

func blend(a, b tcell.Color, t float64) tcell.Color {
	ar, ag, ab := a.RGB()
	br, bg, bb := b.RGB()
	f := func(x, y int32) int32 { return x + int32(float64(y-x)*t) }
	return tcell.NewRGBColor(f(ar, br), f(ag, bg), f(ab, bb))
}

// legacy maps the original 5-color themes (title, background, foreground,
// border, extra) onto roles without changing any color; missing roles are blends.
func legacy(name, title, bg, fg, border, extra string) theme {
	b, s, e := hex(bg), hex(border), hex(extra)
	return theme{name, b, hex(title), hex(title), s, blend(b, s, 0.25), hex(fg), e, blend(e, b, 0.45)}
}

func full(name, bg, main, caret, sub, subAlt, text, err, errExtra string) theme {
	return theme{name, hex(bg), hex(main), hex(caret), hex(sub), hex(subAlt), hex(text), hex(err), hex(errExtra)}
}

var themes = []theme{
	legacy("8008", "#f44c7f", "#333a45", "#e9ecf0", "#939eae", "#da3333"),
	legacy("bento", "#ff7a90", "#2d394d", "#fffaf8", "#4a768d", "#ee2a3a"),
	full("catppuccin", "#1e1e2e", "#cba6f7", "#f5e0dc", "#6c7086", "#313244", "#cdd6f4", "#f38ba8", "#eba0ac"),
	legacy("dark", "#eeeeee", "#111111", "#eeeeee", "#444444", "#da3333"),
	legacy("dots", "#ffffff", "#121520", "#ffffff", "#676e8a", "#791717"),
	legacy("dracula", "#f2f2f2", "#282a36", "#f2f2f2", "#bd93f9", "#ff79c6"),
	legacy("dualshot", "#212222", "#737373", "#212222", "#aaaaaa", "#c82931"),
	full("gruvbox", "#282828", "#fabd2f", "#fabd2f", "#7c6f64", "#3c3836", "#ebdbb2", "#fb4934", "#cc241d"),
	legacy("hammerhead", "#4fcdb9", "#030613", "#e2f1f5", "#1e283a", "#e32b2b"),
	legacy("laser", "#009eaf", "#221b44", "#dbe7e8", "#b82356", "#a8d400"),
	legacy("nautilus", "#ebb723", "#132237", "#1cbaac", "#0b4c6c", "#da3333"),
	full("nord", "#2e3440", "#88c0d0", "#d8dee9", "#616e88", "#3b4252", "#eceff4", "#bf616a", "#d08770"),
	legacy("oblivion", "#a5a096", "#313231", "#f7f5f1", "#5d6263", "#dd452e"),
	legacy("olive", "#92946f", "#e9e5cc", "#373731", "#b7b39e", "#cf2f2f"),
	full("rose pine", "#191724", "#ebbcba", "#ebbcba", "#6e6a86", "#1f1d2e", "#e0def4", "#eb6f92", "#b4637a"),
	full("solarized light", "#fdf6e3", "#268bd2", "#d33682", "#93a1a1", "#eee8d5", "#586e75", "#dc322f", "#cb4b16"),
	full("tokyo night", "#1a1b26", "#7aa2f7", "#c0caf5", "#565f89", "#24283b", "#c0caf5", "#f7768e", "#db4b4b"),
}
