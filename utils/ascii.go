package utils

import "github.com/aldy505/asciitxt"

// Generate ASCII art text from a given string.
func GenerateAscii(s string) string {
	return asciitxt.New(s)
}
