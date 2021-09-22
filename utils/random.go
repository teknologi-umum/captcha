package utils

import (
	"crypto/rand"
	"math/big"
)

func GenerateRandomNumber(limit int) (int, error) {
	randInt, err := rand.Int(rand.Reader, big.NewInt(int64(limit)))
	if err != nil {
		return 0, err
	}

	return int(randInt.Int64()), nil
}
