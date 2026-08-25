package hashutil

import (
	"crypto/hmac"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"golang.org/x/crypto/blake2b"
)

const (
	//  MAC 密钥的最小长度安全策略
	MinMACKeySize = 32
)

// 数据的 BLAKE2b-256 哈希值
func Sum256(data []byte) []byte {
	sum := blake2b.Sum256(data)
	return sum[:]
}

// 数据的 BLAKE2b-256 哈希值并返回十六进制字符串
func Sum256Hex(data []byte) string {
	return hex.EncodeToString(Sum256(data))
}

// 从 io.Reader 流中计算 BLAKE2b-256 哈希值（适用于大文件或流数据）
func Sum256Reader(r io.Reader) ([]byte, error) {
	h, err := blake2b.New256(nil)
	if err != nil {
		return nil, fmt.Errorf("hashutil: failed to create hasher: %w", err)
	}
	if _, err := io.Copy(h, r); err != nil {
		return nil, fmt.Errorf("hashutil: failed to read stream: %w", err)
	}
	return h.Sum(nil), nil
}

// 从 io.Reader 流中计算 BLAKE2b-256 哈希值并返回十六进制字符串
func Sum256ReaderHex(r io.Reader) (string, error) {
	sum, err := Sum256Reader(r)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sum), nil
}

// 将 Token 字符串哈希为十六进制字符串
func HashToken(token string) string {
	return Sum256Hex([]byte(token))
}

// 使用恒定时间比较验证 Token 是否与存储的十六进制哈希匹配
func VerifyToken(token, storedHashHex string) bool {
	stored, err := hex.DecodeString(storedHashHex)
	if err != nil {
		return false
	}
	actual := Sum256([]byte(token))
	return hmac.Equal(actual, stored)
}

// 数据和密钥计算带密钥的 BLAKE2b-256 MAC
func MAC256(data, key []byte) ([]byte, error) {
	if len(key) < MinMACKeySize {
		return nil, fmt.Errorf("hashutil: mac key too short, want >= %d bytes, got %d", MinMACKeySize, len(key))
	}
	h, err := blake2b.New256(key)
	if err != nil {
		return nil, fmt.Errorf("hashutil: invalid mac key: %w", err)
	}
	h.Write(data)
	return h.Sum(nil), nil
}

// 计算带密钥的 BLAKE2b-256 MAC 并返回十六进制字符串
func MAC256Hex(data, key []byte) (string, error) {
	sum, err := MAC256(data, key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sum), nil
}

// 计算带密钥的 BLAKE2b-256 MAC 并返回 URL 安全的 Base64 字符串
func MAC256Base64(data, key []byte) (string, error) {
	sum, err := MAC256(data, key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sum), nil
}

// 从 io.Reader 流中计算带密钥的 BLAKE2b-256 MAC
func MAC256Reader(r io.Reader, key []byte) ([]byte, error) {
	if len(key) < MinMACKeySize {
		return nil, fmt.Errorf("hashutil: mac key too short, want >= %d bytes, got %d", MinMACKeySize, len(key))
	}
	h, err := blake2b.New256(key)
	if err != nil {
		return nil, fmt.Errorf("hashutil: invalid mac key: %w", err)
	}
	if _, err := io.Copy(h, r); err != nil {
		return nil, fmt.Errorf("hashutil: failed to read stream: %w", err)
	}
	return h.Sum(nil), nil
}

// 从 io.Reader 流中计算带密钥的 BLAKE2b-256 MAC 并返回十六进制字符串
func MAC256ReaderHex(r io.Reader, key []byte) (string, error) {
	sum, err := MAC256Reader(r, key)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(sum), nil
}

// 从 io.Reader 流中计算带密钥的 BLAKE2b-256 MAC 并返回 URL 安全的 Base64 字符串
func MAC256ReaderBase64(r io.Reader, key []byte) (string, error) {
	sum, err := MAC256Reader(r, key)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(sum), nil
}

// 验证十六进制编码的 MAC 是否与数据和密钥匹配
func VerifyMAC256Hex(data, key []byte, expectedHex string) bool {
	expected, err := hex.DecodeString(expectedHex)
	if err != nil {
		return false
	}
	actual, err := MAC256(data, key)
	if err != nil {
		return false
	}
	return hmac.Equal(actual, expected)
}

// 验证 Base64 编码的 MAC 是否与数据和密钥匹配
func VerifyMAC256Base64(data, key []byte, expectedBase64 string) bool {
	expected, err := base64.RawURLEncoding.DecodeString(expectedBase64)
	if err != nil {
		return false
	}
	actual, err := MAC256(data, key)
	if err != nil {
		return false
	}
	return hmac.Equal(actual, expected)
}
