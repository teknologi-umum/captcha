package utils_test

import (
	"strings"
	"testing"

	"teknologi-umum-captcha/utils"
)

func TestGenerateAscii(t *testing.T) {
	a := utils.GenerateAscii("Teknologi Umum")
	if !strings.Contains(a, "&lt;") {
		t.Error("GenerateAscii should return ascii")
	}
}
