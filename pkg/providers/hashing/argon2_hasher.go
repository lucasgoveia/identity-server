package hashing

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
)

type Argon2Hasher struct {
}

type Argon2HasherOptions struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

func NewArgon2HasherParams(memory uint32, iterations uint32, parallelism uint8, saltLength uint32, keyLength uint32) *Argon2HasherOptions {
	return &Argon2HasherOptions{
		Memory:      memory,
		Iterations:  iterations,
		Parallelism: parallelism,
		SaltLength:  saltLength,
		KeyLength:   keyLength,
	}
}

func (h *Argon2Hasher) Hash(text string, o ...interface{}) (string, error) {
	options := &Argon2HasherOptions{
		Memory:      64 * 1024, // Default memory
		Iterations:  3,         // Default iterations
		Parallelism: 2,         // Default parallelism
		SaltLength:  16,        // Default salt length
		KeyLength:   32,        // Default key length
	}

	// Check if custom options were provided
	if len(o) > 0 {
		if opts, ok := o[0].(*Argon2HasherOptions); ok {
			options = opts // Use the provided options
		}
	}

	salt, err := generateRandomBytes(options.SaltLength)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(text), salt, options.Iterations, options.Memory, options.Parallelism, options.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, options.Memory, options.Iterations, options.Parallelism, b64Salt, b64Hash)

	return encodedHash, nil
}

func (h *Argon2Hasher) Verify(text, hashed string) (bool, error) {
	// Extract the parameters, salt and derived key from the encoded password hash.
	p, salt, hash, err := decodeHash(hashed)
	if err != nil {
		return false, err
	}

	otherHash := argon2.IDKey([]byte(text), salt, p.Iterations, p.Memory, p.Parallelism, p.KeyLength)

	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return true, nil
	}
	return false, nil
}

func (h *Argon2Hasher) GetAlgorithm() string {
	return "argon2"
}

func generateRandomBytes(n uint32) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

var (
	ErrInvalidHash         = errors.New("the encoded hash is not in the correct format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

func decodeHash(encodedHash string) (p *Argon2HasherOptions, salt, hash []byte, err error) {
	vals := strings.Split(encodedHash, "$")
	if len(vals) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err = fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p = &Argon2HasherOptions{}
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &p.Memory, &p.Iterations, &p.Parallelism)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err = base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.SaltLength = uint32(len(salt))

	hash, err = base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.KeyLength = uint32(len(hash))

	return p, salt, hash, nil
}
