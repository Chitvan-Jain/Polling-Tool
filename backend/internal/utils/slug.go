package utils

import (
	"crypto/rand"
	"math/big"
)

const slugChars = "abcdefghijkmnopqrstuvwxyz23456789"

func GenerateSlug(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(slugChars))))
		if err != nil {
			return "", err
		}
		result[i] = slugChars[n.Int64()]
	}
	return string(result), nil
}