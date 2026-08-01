package utils

import (
	"crypto/rand"
	"math/big"
)

const (
	upperChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerChars = "abcdefghijklmnopqrstuvwxyz"
	digitChars = "0123456789"
	allChars   = upperChars + lowerChars + digitChars
)

// GenerateRandomPassword returns a cryptographically random password of the
// given length containing at least 1 uppercase letter, 1 lowercase letter,
// and 1 digit. If length < 3, it falls back to 12 characters.
func GenerateRandomPassword(length int) string {
	if length < 3 {
		length = 12
	}

	password := make([]byte, 0, length)
	password = append(password, randomChar(upperChars))
	password = append(password, randomChar(lowerChars))
	password = append(password, randomChar(digitChars))

	for i := 3; i < length; i++ {
		password = append(password, randomChar(allChars))
	}

	// Fisher-Yates shuffle to avoid a predictable prefix pattern
	for i := len(password) - 1; i > 0; i-- {
		j := cryptoRandInt(i + 1)
		password[i], password[j] = password[j], password[i]
	}

	return string(password)
}

func randomChar(charset string) byte {
	return charset[cryptoRandInt(len(charset))]
}

func cryptoRandInt(max int) int {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(max)))
	if err != nil {
		// crypto/rand failure is practically impossible; never block the app
		return 0
	}
	return int(n.Int64())
}
