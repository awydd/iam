package random

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
)

const (
	AlphanumericCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	UnambiguousCharset  = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	NumericCharset      = "0123456789"
)

var errNonPositiveLength = errors.New("random: length must be positive")

func Alphanumeric(length int) (string, error) {
	return FromCharset(length, AlphanumericCharset)
}

func Unambiguous(length int) (string, error) {
	return FromCharset(length, UnambiguousCharset)
}

func Numeric(length int) (string, error) {
	return FromCharset(length, NumericCharset)
}

func FromCharset(length int, charset string) (string, error) {
	if length <= 0 {
		return "", errNonPositiveLength
	}
	if len(charset) == 0 {
		return "", errors.New("random: charset must not be empty")
	}

	charsetLen := big.NewInt(int64(len(charset)))
	result := make([]byte, length)
	for i := range result {
		idx, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", fmt.Errorf("random: generate random char: %w", err)
		}
		// 指定的 charset 下标的 ascii 值
		result[i] = charset[idx.Int64()]
	}

	return string(result), nil
}

func Bytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, errNonPositiveLength
	}
	buf := make([]byte, length)
	// 没有报错就随机生成了，Read 方法自动随机填充
	if _, err := rand.Read(buf); err != nil {
		return nil, fmt.Errorf("random: read random bytes: %w", err)
	}
	return buf, nil
}

// 这里生成的字符串 len() 是 length 的 1.33x 倍，6 -> 8
func Token(length int) (string, error) {
	b, err := Bytes(length)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// 这里生成的字符串 len() 是 length 的两倍，6 -> 12
func Hex(length int) (string, error) {
	b, err := Bytes(length)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
