package shortener

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

const (
	alphabet   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	codeLength = 6
	MaxRetries = 10
)

func GenerateCode() (string, error) {
	b := make([]byte, codeLength)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random code: %w", err)
		}
		b[i] = alphabet[n.Int64()]
	}
	return string(b), nil
}

type ErrMaxRetries struct{}

func (e ErrMaxRetries) Error() string {
	return "unable to generate a unique short code after maximum retries"
}
