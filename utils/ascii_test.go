package utils_test

import (
	"strings"
	"teknologi-umum-bot/utils"
	"testing"
)

func TestGenerateAscii(t *testing.T) {
	a := utils.GenerateAscii("Teknologi Umum")
	if !strings.Contains(a, "&lt;") {
		t.Error("GenerateAscii should return ascii")
	}
}
