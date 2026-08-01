package utils

import (
	"strings"
	"testing"
)

func TestGenerateRandomPassword(t *testing.T) {
	tests := []struct {
		name   string
		length int
	}{
		{"default length fallback", 1},
		{"exact length", 12},
		{"long password", 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pw := GenerateRandomPassword(tt.length)

			expected := tt.length
			if tt.length < 3 {
				expected = 12
			}
			if len(pw) != expected {
				t.Errorf("GenerateRandomPassword(%d) length = %d, want %d", tt.length, len(pw), expected)
			}

			if !strings.ContainsAny(pw, upperChars) {
				t.Errorf("password %q missing uppercase letter", pw)
			}
			if !strings.ContainsAny(pw, lowerChars) {
				t.Errorf("password %q missing lowercase letter", pw)
			}
			if !strings.ContainsAny(pw, digitChars) {
				t.Errorf("password %q missing digit", pw)
			}
		})
	}
}

func TestGenerateRandomPasswordUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 50; i++ {
		pw := GenerateRandomPassword(12)
		if seen[pw] {
			t.Fatalf("duplicate password generated: %q", pw)
		}
		seen[pw] = true
	}
}
