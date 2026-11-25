// Package data embeds word lists so the binary runs from any directory.
package data

import _ "embed"

//go:embed words.json
var Words []byte

//go:embed quotes.json
var Quotes []byte
