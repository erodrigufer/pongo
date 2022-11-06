package sysutils

import (
	"crypto/rand"
	"fmt"
	"math/big"
)

// NewRandomString, returns a random string.
// Parameters: length is the number of characters of the string that should be
// returned. charset, is the valid character set from which to generate a
// random string.
func NewRandomString(length int, charset string) (string, error) {
	// Make a slice of length length, in which to store random characters.
	b := make([]byte, length)
	for i := range b {
		r, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		r64 := r.Int64()
		if err != nil {
			return "", fmt.Errorf("error: getting a random integer failed: %w", err)
		}
		// Pick a single character from the character-set through indexing the
		// string of the character-set. The index is a random number between 0
		// and the length of the character-set minus 1.
		b[i] = charset[int(r64)]
	}

	return string(b), nil
}

// NewRandomPassword, returns a random password of given character length.
func NewRandomPassword(length int, charset string) (string, error) {
	s, err := NewRandomString(length, charset)
	if err != nil {
		return "", fmt.Errorf("error: could not create random password: %w", err)
	}
	return s, nil
}

// NewRandomUsername, returns a random password of the given character
// length.
func NewRandomUsername(length int, charset string) (string, error) {
	s, err := NewRandomString(length, charset)
	if err != nil {
		return "", fmt.Errorf("error: could not create random username: %w", err)
	}
	return s, nil
}
