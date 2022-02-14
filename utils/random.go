package utils

import (
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// GenerateRandomNumber generates a random number from 000 to 999
func GenerateRandomNumber() string {
	rand.Seed(time.Now().UnixMilli())
	var out strings.Builder
	for i := 0; i < 3; i++ {
		out.WriteString(strconv.Itoa(rand.Intn(9)))
	}

	return out.String()
}
