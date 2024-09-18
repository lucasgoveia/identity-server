package security

import (
	"crypto/rand"
	"math/big"
)

type SecureKeyGenerator struct{}

func NewSecureKeyGenerator() *SecureKeyGenerator {
	return &SecureKeyGenerator{}
}

func (g *SecureKeyGenerator) Generate(alphabet []rune, keyLength int) (string, error) {
	result := make([]rune, keyLength)
	for i := range result {
		index, err := rand.Int(rand.Reader, big.NewInt(int64(len(alphabet))))
		if err != nil {
			return "", err
		}
		result[i] = alphabet[index.Int64()]
	}
	return string(result), nil
}
