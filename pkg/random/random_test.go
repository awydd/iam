package random

import (
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"strings"
	"testing"
)

func TestFromCharset(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		charset string
		wantErr bool
	}{
		{"valid alphanumeric", 16, AlphanumericCharset, false},
		{"valid unambiguous", 10, UnambiguousCharset, false},
		{"valid numeric", 6, NumericCharset, false},
		{"length one", 1, AlphanumericCharset, false},
		{"zero length", 0, AlphanumericCharset, true},
		{"negative length", -1, AlphanumericCharset, true},
		{"empty charset", 10, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromCharset(tt.length, tt.charset)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (result=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != tt.length {
				t.Fatalf("length = %d, want %d", len(got), tt.length)
			}
			for _, c := range got {
				if !strings.ContainsRune(tt.charset, c) {
					t.Fatalf("result contains char %q not in charset %q", c, tt.charset)
				}
			}
		})
	}
}

func TestAlphanumeric(t *testing.T) {
	s, err := Alphanumeric(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) != 20 {
		t.Fatalf("length = %d, want 20", len(s))
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(s) {
		t.Fatalf("result %q contains unexpected characters", s)
	}
}

func TestUnambiguous(t *testing.T) {
	s, err := Unambiguous(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range []rune{'0', 'O', '1', 'l', 'I'} {
		if strings.ContainsRune(s, c) {
			t.Fatalf("result %q contains ambiguous character %q", s, c)
		}
	}
}

func TestNumeric(t *testing.T) {
	s, err := Numeric(6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !regexp.MustCompile(`^[0-9]{6}$`).MatchString(s) {
		t.Fatalf("result %q is not a 6-digit numeric string", s)
	}
}

func TestBytes(t *testing.T) {
	t.Run("valid length", func(t *testing.T) {
		b, err := Bytes(32)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(b) != 32 {
			t.Fatalf("length = %d, want 32", len(b))
		}
	})

	t.Run("zero length", func(t *testing.T) {
		if _, err := Bytes(0); err == nil {
			t.Fatal("expected error for zero length")
		}
	})

	t.Run("negative length", func(t *testing.T) {
		if _, err := Bytes(-5); err == nil {
			t.Fatal("expected error for negative length")
		}
	})
}

func TestToken(t *testing.T) {
	s, err := Token(24)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	decoded, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("token is not valid base64url: %v", err)
	}
	if len(decoded) != 24 {
		t.Fatalf("decoded length = %d, want 24", len(decoded))
	}
	if strings.ContainsAny(s, "+/=") {
		t.Fatalf("token %q contains non-URL-safe characters", s)
	}
}

func TestHex(t *testing.T) {
	s, err := Hex(16)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s) != 32 {
		t.Fatalf("length = %d, want 32", len(s))
	}
	if _, err := hex.DecodeString(s); err != nil {
		t.Fatalf("result %q is not valid hex: %v", s, err)
	}
}

func TestUniqueness(t *testing.T) {
	const n = 1000
	seen := make(map[string]bool, n)

	for range n {
		s, err := Alphanumeric(16)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[s] {
			t.Fatalf("duplicate value generated: %q", s)
		}
		seen[s] = true
	}
}
