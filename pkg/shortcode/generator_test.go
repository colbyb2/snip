package shortcode

import (
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	g := NewGenerator(DefaultLength)

	code, err := g.Generate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(code) != DefaultLength {
		t.Errorf("expected length %d, got %d", DefaultLength, len(code))
	}

	// Verify all characters are from the alphabet
	for _, c := range code {
		found := false
		for _, a := range alphabet {
			if c == a {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("character %c not in alphabet", c)
		}
	}
}

func TestGenerator_Generate_Uniqueness(t *testing.T) {
	g := NewGenerator(DefaultLength)
	seen := make(map[string]bool)
	iterations := 10000

	for i := 0; i < iterations; i++ {
		code, err := g.Generate()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[code] {
			t.Errorf("duplicate code generated: %s", code)
		}
		seen[code] = true
	}
}

func TestGenerator_CustomLength(t *testing.T) {
	tests := []struct {
		name     string
		length   int
		expected int
	}{
		{"default on zero", 0, DefaultLength},
		{"default on negative", -1, DefaultLength},
		{"custom length 5", 5, 5},
		{"custom length 10", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGenerator(tt.length)
			if g.Length() != tt.expected {
				t.Errorf("expected length %d, got %d", tt.expected, g.Length())
			}

			code, err := g.Generate()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(code) != tt.expected {
				t.Errorf("expected code length %d, got %d", tt.expected, len(code))
			}
		})
	}
}

func TestGenerator_PossibleCombinations(t *testing.T) {
	g := NewGenerator(7)
	combinations := g.PossibleCombinations()

	// 55^7 = 1,152,921,504,606,846,975 (but we use int64 so it caps)
	// Actually 55^7 = 1,522,070,312,500 (wrong, let me recalc)
	// 55^1 = 55
	// 55^2 = 3025
	// 55^3 = 166375
	// 55^7 ≈ 1.5 trillion
	
	if combinations <= 0 {
		t.Errorf("expected positive combinations, got %d", combinations)
	}

	// With 7 characters and 55-char alphabet, should be > 1 trillion
	if combinations < 1_000_000_000_000 {
		t.Errorf("expected > 1 trillion combinations, got %d", combinations)
	}
}

func BenchmarkGenerator_Generate(b *testing.B) {
	g := NewGenerator(DefaultLength)
	
	for i := 0; i < b.N; i++ {
		_, _ = g.Generate()
	}
}
