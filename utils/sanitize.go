package utils

import "strings"

func SanitizeInput(inp string) string {
	var str string
	str = strings.ReplaceAll(inp, ">", "&gt;")
	str = strings.ReplaceAll(str, "<", "&lt;")
	return str
}
