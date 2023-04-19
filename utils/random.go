package utils

import (
	"math/rand"
	"strconv"
	"strings"
)

// GenerateRandomNumber generates a random sequence string
func GenerateRandomNumber() string {
	var out strings.Builder
	for i := 0; i < 3; i++ {
		randomNumber := rand.Intn(14)
		if randomNumber == 10 {
			out.WriteString("V")
		} else if randomNumber == 11 {
			out.WriteString("W")
		} else if randomNumber == 12 {
			out.WriteString("X")
		} else if randomNumber == 13 {
			out.WriteString("Y")
		} else {
			out.WriteString(strconv.Itoa(randomNumber))
		}
	}

	return out.String()
}
