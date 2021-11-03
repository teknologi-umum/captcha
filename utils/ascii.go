package utils

import (
	"strings"

	"github.com/aldy505/asciitxt"
)

// Generate ASCII art text from a given string.
func GenerateAscii(s string) string {
	text := asciitxt.New(s)
	// then we need to sanitize it
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	return text
}
