package utils

import (
	"crypto/rand"
	"math/big"
)

const base62Alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func GenerateBase62Key(length int) (string, error) {
	b := make([]byte, length)
	alphabetLength := big.NewInt(int64(len(base62Alphabet)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, alphabetLength)
		if err != nil {
			return "", err
		}
		b[i] = base62Alphabet[num.Int64()]
	}

	return string(b), nil
}
