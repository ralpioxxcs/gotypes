# gotypes

[![](https://goreportcard.com/badge/github.com/ralpioxxcs/gotypes)](https://goreportcard.com/report/github.com/ralpioxxcs/gotypes)
[![Actions Status: build](https://github.com/ralpioxxcs/gotypes/workflows/build/badge.svg)](https://github.com/ralpioxxcs/gotypes/actions?query=workflow%3A"build")

terminal based typing test application using golang inspirated by [monkeytype](https://monkeytype.com/)

![Screenshot](gotypes.gif)

## Features
* words (15/30/60/120) and time (15/30/60s) modes
* paragraph or tape (single scrolling line) text
* result screen: wpm, accuracy, raw, consistency, per-second chart
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
| <kbd>esc</kbd> | menu: mode, text style, theme (themes preview on hover) |
| <kbd>ctrl</kbd>+<kbd>w</kbd> / <kbd>alt</kbd>+<kbd>backspace</kbd> | delete word |
| <kbd>←</kbd> <kbd>→</kbd> or mouse | inspect a second on the result chart |
| <kbd>ctrl</kbd>+<kbd>c</kbd> | quit |

A truecolor terminal is recommended for accurate theme colors.

## Docker
```
docker run -it --rm -e "TERM=xterm-256color" placket/gotypes:latest
```

## Dependencies
* https://github.com/rivo/tview

## License
See [LICENSE](LICENSE) file
