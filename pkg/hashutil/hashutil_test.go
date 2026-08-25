package hashutil

import (
	"bytes"
	"strings"
	"testing"
)

func TestSum256(t *testing.T) {
	data := []byte("hello world")
	sum := Sum256(data)

	if len(sum) != 32 {
		t.Fatalf("Sum256 length = %d, want 32", len(sum))
	}

	// deterministic: same input -> same output
	sum2 := Sum256(data)
	if !bytes.Equal(sum, sum2) {
		t.Fatal("Sum256 is not deterministic for the same input")
	}

	// different input -> different output
	sum3 := Sum256([]byte("hello world!"))
	if bytes.Equal(sum, sum3) {
		t.Fatal("Sum256 produced the same output for different inputs")
	}
}

func TestSum256Hex(t *testing.T) {
	got := Sum256Hex([]byte("hello"))
	if len(got) != 64 { // 32 bytes -> 64 hex chars
		t.Fatalf("length = %d, want 64", len(got))
	}
	if strings.ToLower(got) != got {
		t.Fatalf("expected lowercase hex, got %q", got)
	}
}

func TestHashTokenAndVerifyToken(t *testing.T) {
	token := "some-refresh-token-value"
	hashed := HashToken(token)

	if !VerifyToken(token, hashed) {
		t.Fatal("VerifyToken failed to verify a correctly hashed token")
	}

	if VerifyToken("wrong-token", hashed) {
		t.Fatal("VerifyToken incorrectly verified a mismatched token")
	}
}

func TestVerifyToken_InvalidHex(t *testing.T) {
	if VerifyToken("some-token", "not-valid-hex-zzz") {
		t.Fatal("VerifyToken should return false for malformed hex input")
	}
}

func TestVerifyToken_EmptyStoredHash(t *testing.T) {
	if VerifyToken("some-token", "") {
		t.Fatal("VerifyToken should return false when stored hash is empty")
	}
}

func TestMAC256(t *testing.T) {
	key := make([]byte, MinMACKeySize)
	for i := range key {
		key[i] = byte(i)
	}
	data := []byte("payload")

	sum, err := MAC256(data, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sum) != 32 {
		t.Fatalf("MAC length = %d, want 32", len(sum))
	}

	// deterministic for the same key+data
	sum2, err := MAC256(data, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !bytes.Equal(sum, sum2) {
		t.Fatal("MAC256 is not deterministic for the same key and data")
	}

	// different key -> different MAC
	otherKey := make([]byte, MinMACKeySize)
	copy(otherKey, key)
	otherKey[0] ^= 0xFF
	sum3, err := MAC256(data, otherKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bytes.Equal(sum, sum3) {
		t.Fatal("MAC256 produced the same output for different keys")
	}
}

func TestMAC256_KeyTooShort(t *testing.T) {
	shortKey := make([]byte, MinMACKeySize-1)
	_, err := MAC256([]byte("data"), shortKey)
	if err == nil {
		t.Fatal("expected error for key shorter than MinMACKeySize, got nil")
	}
}

func TestMAC256_EmptyKey(t *testing.T) {
	_, err := MAC256([]byte("data"), nil)
	if err == nil {
		t.Fatal("expected error for empty key, got nil")
	}
}

func TestMAC256Hex_And_VerifyMAC256Hex(t *testing.T) {
	key := bytes.Repeat([]byte{0xAB}, MinMACKeySize)
	data := []byte("important payload")

	macHex, err := MAC256Hex(data, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(macHex) != 64 {
		t.Fatalf("length = %d, want 64", len(macHex))
	}

	if !VerifyMAC256Hex(data, key, macHex) {
		t.Fatal("VerifyMAC256Hex failed to verify a valid MAC")
	}
	if VerifyMAC256Hex([]byte("tampered payload"), key, macHex) {
		t.Fatal("VerifyMAC256Hex incorrectly verified a MAC for different data")
	}
	if VerifyMAC256Hex(data, key, "deadbeef") {
		t.Fatal("VerifyMAC256Hex incorrectly verified a wrong MAC")
	}
}

func TestMAC256Hex_KeyTooShort(t *testing.T) {
	shortKey := bytes.Repeat([]byte{0x01}, MinMACKeySize-1)
	if _, err := MAC256Hex([]byte("data"), shortKey); err == nil {
		t.Fatal("expected error for short key, got nil")
	}
}

func TestVerifyMAC256Hex_InvalidHexInput(t *testing.T) {
	key := bytes.Repeat([]byte{0x01}, MinMACKeySize)
	if VerifyMAC256Hex([]byte("data"), key, "not-hex-zz") {
		t.Fatal("expected false for malformed hex expected value")
	}
}

func TestMAC256Base64_And_VerifyMAC256Base64(t *testing.T) {
	key := bytes.Repeat([]byte{0xCD}, MinMACKeySize)
	data := []byte("important payload")

	macB64, err := MAC256Base64(data, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.ContainsAny(macB64, "+/=") {
		t.Fatalf("expected URL-safe base64 without padding, got %q", macB64)
	}

	if !VerifyMAC256Base64(data, key, macB64) {
		t.Fatal("VerifyMAC256Base64 failed to verify a valid MAC")
	}
	if VerifyMAC256Base64([]byte("tampered"), key, macB64) {
		t.Fatal("VerifyMAC256Base64 incorrectly verified a MAC for different data")
	}
}

func TestVerifyMAC256Base64_InvalidInput(t *testing.T) {
	key := bytes.Repeat([]byte{0x01}, MinMACKeySize)
	if VerifyMAC256Base64([]byte("data"), key, "not base64!!") {
		t.Fatal("expected false for malformed base64 expected value")
	}
}

func TestMAC256Base64_KeyTooShort(t *testing.T) {
	shortKey := bytes.Repeat([]byte{0x01}, MinMACKeySize-1)
	if _, err := MAC256Base64([]byte("data"), shortKey); err == nil {
		t.Fatal("expected error for short key, got nil")
	}
}
