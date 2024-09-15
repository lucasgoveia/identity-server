package hashing_test

import (
	"identity-server/pkg/providers/hashing"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArgon2Hasher_HashAndVerify(t *testing.T) {
	hasher := &hashing.Argon2Hasher{}

	t.Run("Hash and Verify with default options", func(t *testing.T) {
		text := "password123"
		hashed, err := hasher.Hash(text)
		assert.NoError(t, err, "Hashing should not return an error")
		assert.True(t, strings.Contains(hashed, "$argon2id$"), "Hashed string should contain Argon2id prefix")

		isValid, err := hasher.Verify(text, hashed)
		assert.NoError(t, err, "Verifying should not return an error")
		assert.True(t, isValid, "The original text should match the hash")
	})

	t.Run("Verify with wrong password", func(t *testing.T) {
		text := "password123"
		hashed, err := hasher.Hash(text)
		assert.NoError(t, err)

		isValid, err := hasher.Verify("wrongpassword", hashed)
		assert.NoError(t, err)
		assert.False(t, isValid, "Verification should fail with incorrect password")
	})

	t.Run("Invalid hash format", func(t *testing.T) {
		_, err := hasher.Verify("password", "invalidhash")
		assert.ErrorIs(t, err, hashing.ErrInvalidHash, "Invalid hash format should return ErrInvalidHash")
	})

	t.Run("Incompatible Argon2 version", func(t *testing.T) {
		incompatibleHash := "$argon2id$v=99$m=65536,t=3,p=2$YXJnb24$YXJnb24=" // A fake hash with an incompatible version
		_, err := hasher.Verify("password", incompatibleHash)
		assert.ErrorIs(t, err, hashing.ErrIncompatibleVersion, "Incompatible version should return ErrIncompatibleVersion")
	})

	t.Run("Custom Argon2 options", func(t *testing.T) {
		options := hashing.NewArgon2HasherParams(128*1024, 4, 1, 32, 64)
		text := "customoptions"
		hashed, err := hasher.Hash(text, options)
		assert.NoError(t, err, "Hashing should not return an error with custom options")
		assert.True(t, strings.Contains(hashed, "$argon2id$"), "Hashed string should contain Argon2id prefix")

		isValid, err := hasher.Verify(text, hashed)
		assert.NoError(t, err, "Verifying should not return an error")
		assert.True(t, isValid, "The original text should match the hash with custom options")
	})
}

func TestArgon2Hasher_GetAlgorithm(t *testing.T) {
	hasher := &hashing.Argon2Hasher{}
	algorithm := hasher.GetAlgorithm()
	assert.Equal(t, "argon2", algorithm, "GetAlgorithm should return 'argon2'")
}

func TestArgon2Hasher_EmptyString(t *testing.T) {
	hasher := &hashing.Argon2Hasher{}

	t.Run("Hash empty string", func(t *testing.T) {
		hashed, err := hasher.Hash("")
		assert.NoError(t, err, "Hashing empty string should not return an error")
		assert.True(t, strings.Contains(hashed, "$argon2id$"), "Hashed empty string should contain Argon2id prefix")

		isValid, err := hasher.Verify("", hashed)
		assert.NoError(t, err, "Verifying empty string should not return an error")
		assert.True(t, isValid, "Empty string should match the empty string hash")
	})

	t.Run("Verify empty string with non-empty hash", func(t *testing.T) {
		nonEmptyHashed, err := hasher.Hash("nonEmpty")
		assert.NoError(t, err)

		isValid, err := hasher.Verify("", nonEmptyHashed)
		assert.NoError(t, err)
		assert.False(t, isValid, "Empty string should not match a non-empty string hash")
	})
}

func TestArgon2Hasher_InvalidCustomOptions(t *testing.T) {
	hasher := &hashing.Argon2Hasher{}

	t.Run("Invalid custom options: zero iterations", func(t *testing.T) {
		options := hashing.NewArgon2HasherParams(64*1024, 0, 2, 16, 32) // Invalid: iterations cannot be zero

		_, err := hasher.Hash("test", options)
		assert.Error(t, err, "Hashing should fail with invalid options (iterations cannot be zero)")
	})

	t.Run("Invalid custom options: zero memory", func(t *testing.T) {
		options := hashing.NewArgon2HasherParams(0, 3, 2, 16, 32) // Invalid: memory cannot be zero

		_, err := hasher.Hash("test", options)
		assert.Error(t, err, "Hashing should fail with invalid options (memory cannot be zero)")
	})

	t.Run("Invalid custom options: zero parallelism", func(t *testing.T) {
		options := hashing.NewArgon2HasherParams(64*1024, 3, 0, 16, 32) // Invalid: parallelism cannot be zero

		_, err := hasher.Hash("test", options)
		assert.Error(t, err, "Hashing should fail with invalid options (parallelism cannot be zero)")
	})
}

func TestArgon2Hasher_VeryLargeInput(t *testing.T) {
	hasher := &hashing.Argon2Hasher{}

	t.Run("Hash and Verify very large string", func(t *testing.T) {
		// Create a very large input string (e.g., 1 million characters)
		largeInput := strings.Repeat("a", 1_000_000)

		hashed, err := hasher.Hash(largeInput)
		assert.NoError(t, err, "Hashing large input should not return an error")

		isValid, err := hasher.Verify(largeInput, hashed)
		assert.NoError(t, err, "Verifying large input should not return an error")
		assert.True(t, isValid, "The large input should match the hash")
	})
}

func TestArgon2Hasher_BoundaryOptions(t *testing.T) {
	hasher := &hashing.Argon2Hasher{}

	t.Run("Hash and Verify with boundary memory value", func(t *testing.T) {
		options := hashing.NewArgon2HasherParams(8*1024, 3, 2, 16, 32) // Minimal valid memory for Argon2

		text := "test_boundary_memory"
		hashed, err := hasher.Hash(text, options)
		assert.NoError(t, err, "Hashing should succeed with boundary memory value")

		isValid, err := hasher.Verify(text, hashed)
		assert.NoError(t, err, "Verifying should not return an error")
		assert.True(t, isValid, "The original text should match the hash with boundary memory value")
	})

	t.Run("Hash and Verify with boundary iteration value", func(t *testing.T) {
		options := hashing.NewArgon2HasherParams(64*1024, 1, 2, 16, 32) // Minimal valid iterations for Argon2

		text := "test_boundary_iterations"
		hashed, err := hasher.Hash(text, options)
		assert.NoError(t, err, "Hashing should succeed with boundary iteration value")

		isValid, err := hasher.Verify(text, hashed)
		assert.NoError(t, err, "Verifying should not return an error")
		assert.True(t, isValid, "The original text should match the hash with boundary iteration value")
	})
}

func TestArgon2Hasher_InvalidHash(t *testing.T) {
	hasher := &hashing.Argon2Hasher{}

	t.Run("Invalid Argon2 hash format", func(t *testing.T) {
		invalidHash := "$argon2id$v=1$m=65536,t=3,p=2$invalidsalt$invalidhash"

		_, err := hasher.Verify("test", invalidHash)
		assert.ErrorIs(t, err, hashing.ErrIncompatibleVersion, "Verifying with an invalid hash format should return ErrInvalidHash")
	})

	t.Run("Incompatible version with valid hash structure", func(t *testing.T) {
		incompatibleHash := "$argon2id$v=99$m=65536,t=3,p=2$YXJnb24$YXJnb24="

		_, err := hasher.Verify("test", incompatibleHash)
		assert.ErrorIs(t, err, hashing.ErrIncompatibleVersion, "Verifying with incompatible version should return ErrIncompatibleVersion")
	})
}
