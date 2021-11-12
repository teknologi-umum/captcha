package utils_test

import (
	"teknologi-umum-bot/utils"
	"testing"
)

func TestIsIn(t *testing.T) {
	i := utils.IsIn([]string{"a", "b", "c"}, "a")
	if i != true {
		t.Error("Expected true, got false")
	}

	i = utils.IsIn([]string{"a", "b", "c"}, "d")
	if i != false {
		t.Error("Expected false, got true")
	}
}
