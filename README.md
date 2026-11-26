# gotypes

[![](https://goreportcard.com/badge/github.com/ralpioxxcs/gotypes)](https://goreportcard.com/report/github.com/ralpioxxcs/gotypes)
[![Actions Status: build](https://github.com/ralpioxxcs/gotypes/workflows/build/badge.svg)](https://github.com/ralpioxxcs/gotypes/actions?query=workflow%3A"build")

terminal based typing test application using golang inspirated by [monkeytype](https://monkeytype.com/)

![Screenshot](gotypes.gif)

## Features
* words (15/30/60/120), time (15/30/60s) and quote (short/medium/long) modes
* paragraph or tape (single scrolling line) text
* live speed gauge while typing (last 3s, 타/분 for korean) with peak hold
* result screen: wpm, accuracy, raw, consistency, per-second chart
* languages: english, english1000, korean (also shows 타/분), polish
* settings and results are saved; the result screen shows your personal best per mode and language
  (`config.json` and `history.jsonl` in your OS config dir, e.g. `~/.config/gotypes` or `~/Library/Application Support/gotypes`)
* [gmk](https://www.gmk-electronic-design.de/en/products/keycaps) keycap themes plus nord, gruvbox, catppuccin, tokyo night, rose pine, solarized light

## Install
```
$ go install github.com/ralpioxxcs/gotypes@latest
```

## Usage
Run `gotypes` (or `go run .` in the source root) and start typing.

| key | action |
|---|---|
| <kbd>tab</kbd> | restart |
| <kbd>esc</kbd> | settings: mode, text, language, theme. Enter a category, type to filter, enter to apply (themes preview as you move) |
| <kbd>ctrl</kbd>+<kbd>w</kbd> / <kbd>alt</kbd>+<kbd>backspace</kbd> | delete word |
| <kbd>←</kbd> <kbd>→</kbd> or mouse | inspect a second on the result chart |
| click | pick a mode on the mode bar (`words 15 30 …`) |
| <kbd>ctrl</kbd>+<kbd>c</kbd> | quit |

A truecolor terminal is recommended for accurate theme colors.

English quotes are excerpts of public-domain works taken verbatim from [Project Gutenberg](https://www.gutenberg.org), with curly quotes and dashes written as plain ASCII so they can be typed.
Korean quotes are proverbs and excerpts of public-domain works (김유정, 김소월, 나도향, 이효석, 현진건) taken verbatim from [Korean Wikisource](https://ko.wikisource.org).

## Play over SSH
Host it, and anyone can play with just `ssh`:
```
$ gotypes -ssh :2222 -hostkey ./gotypes_ed25519 -data ./data
$ ssh -p 2222 your.host        # from anywhere
```
* **Race**: pick `race` in the mode menu to join the room for your language (up to 8 players), then press enter when ready. A 10 s countdown starts once two players are ready; anyone not ready by then moves on to the next room. Everyone types the same quote with live progress lanes. Your ssh user name is your racer name (`ssh name@your.host`).
* Connecting with a public key keeps that key's settings and history under `-data`; without one you play unsaved.
* Only interactive game sessions are served: commands, sessions without a PTY and port forwarding are refused.
* Limits: `-max-sessions` (default 50), per-IP connection rate limit, 10 min idle timeout, 2 h max session.
* Keep the host key file, or clients will warn that the host key changed.
* Colors are 24-bit by default; on a terminal without truecolor use `ssh -o SetEnv=TCELL_TRUECOLOR=disable ...`.

## Docker
```
# play
docker run -it --rm -e "TERM=xterm-256color" placket/gotypes:latest
# serve over SSH (host key and player data in the volume)
docker run -d -p 2222:2222 -v gotypes:/data placket/gotypes:latest -ssh :2222 -hostkey /data/hostkey -data /data
```

## Dependencies
* https://github.com/gdamore/tcell
* https://github.com/charmbracelet/wish (SSH server)

## License
See [LICENSE](LICENSE) file
