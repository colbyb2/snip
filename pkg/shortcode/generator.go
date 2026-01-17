// Package shortcode provides utilities for generating short URL codes.
package shortcode

import (
	"crypto/rand"
	"math/big"
)

// alphabet contains characters used for short codes.
// Excludes ambiguous characters (0, O, l, 1, I) for readability.
const alphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz"

// DefaultLength is the default length for generated short codes.
const DefaultLength = 7

// Generator creates unique short codes.
type Generator struct {
	length int
}

// NewGenerator creates a new Generator with the specified code length.
func NewGenerator(length int) *Generator {
	if length <= 0 {
		length = DefaultLength
	}
	return &Generator{length: length}
}

// Generate creates a new random short code.
// Uses crypto/rand for secure randomness.
func (g *Generator) Generate() (string, error) {
	result := make([]byte, g.length)
	alphabetLen := big.NewInt(int64(len(alphabet)))

	for i := 0; i < g.length; i++ {
		num, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		result[i] = alphabet[num.Int64()]
	}

	return string(result), nil
}

// Length returns the configured code length.
func (g *Generator) Length() int {
	return g.length
}

// PossibleCombinations returns the number of possible unique codes.
// With default settings (7 chars, 55 char alphabet): ~1.1 trillion combinations
func (g *Generator) PossibleCombinations() int64 {
	result := int64(1)
	for i := 0; i < g.length; i++ {
		result *= int64(len(alphabet))
	}
	return result
}
