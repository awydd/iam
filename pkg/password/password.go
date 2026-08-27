package password

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/argon2"
)

const (
	defaultVariant = "argon2id"
	defaultVersion = argon2.Version

	maxPasswordLength = 256

	maxDecodedMemory  = 1 << 20 // 1 GiB
	maxDecodedTime    = 100
	maxDecodedThreads = 64
)

var (
	ErrEmptyPassword   = errors.New("password cannot be empty")
	ErrPasswordTooLong = errors.New("password exceeds maximum allowed length")
	ErrBusy            = errors.New("server too busy")
	ErrInvalidHash     = errors.New("invalid or corrupted hash format")
)

type Config struct {
	Time, Memory       uint32
	Threads            uint8
	KeyLength, SaltLen uint32
	MaxConcurrency     int
	WaitTimeout        time.Duration
}

func DefaultConfig() *Config {
	return &Config{
		Time: 3, Memory: 1 << 15, Threads: 2, // 32MB Memory
		KeyLength: 32, SaltLen: 16,
		MaxConcurrency: 8, WaitTimeout: 5 * time.Second,
	}
}

func (cfg *Config) withDefaults() *Config {
	def := DefaultConfig()
	if cfg == nil {
		return def
	}
	out := *cfg
	if out.Time == 0 {
		out.Time = def.Time
	}
	if out.Memory == 0 {
		out.Memory = def.Memory
	}
	if out.Threads == 0 {
		out.Threads = def.Threads
	}
	if out.KeyLength == 0 {
		out.KeyLength = def.KeyLength
	}
	if out.SaltLen == 0 {
		out.SaltLen = def.SaltLen
	}
	if out.MaxConcurrency <= 0 {
		out.MaxConcurrency = def.MaxConcurrency
	}
	if out.WaitTimeout <= 0 {
		out.WaitTimeout = def.WaitTimeout
	}
	return &out
}

type hasher struct {
	config    *Config
	semaphore chan struct{}
}

func newHasher(cfg *Config) *hasher {
	cfg = cfg.withDefaults()
	return &hasher{
		config:    cfg,
		semaphore: make(chan struct{}, cfg.MaxConcurrency),
	}
}

var (
	defaultHasher *hasher
	once          sync.Once
)

var globalConfig *Config

func Init(cfg *Config) {
	globalConfig = cfg
}

func getHasher() *hasher {
	once.Do(func() {
		defaultHasher = newHasher(globalConfig)
	})

	if defaultHasher == nil {
		defaultHasher = newHasher(nil)
	}
	return defaultHasher
}

func Hash(password string) (string, error) {
	return getHasher().hashWithContext(context.Background(), password)
}

func HashWithContext(ctx context.Context, password string) (string, error) {
	return getHasher().hashWithContext(ctx, password)
}

func Validate(password, encodedHash string) (bool, error) {
	return getHasher().validateWithContext(context.Background(), password, encodedHash)
}

func ValidateWithContext(ctx context.Context, password, encodedHash string) (bool, error) {
	return getHasher().validateWithContext(ctx, password, encodedHash)
}

func NeedsRehash(encodedHash string) bool {
	return getHasher().needsRehash(encodedHash)
}

func (h *hasher) hashWithContext(ctx context.Context, password string) (string, error) {
	if password == "" {
		return "", ErrEmptyPassword
	}
	if len(password) > maxPasswordLength {
		return "", ErrPasswordTooLong
	}

	if err := h.acquire(ctx); err != nil {
		return "", err
	}
	defer h.release()

	salt := make([]byte, h.config.SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.config.Time,
		h.config.Memory,
		h.config.Threads,
		h.config.KeyLength,
	)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$%s$v=%d$m=%d,t=%d,p=%d$%s$%s",
		defaultVariant, defaultVersion, h.config.Memory, h.config.Time, h.config.Threads, b64Salt, b64Hash), nil
}

func (h *hasher) validateWithContext(ctx context.Context, password string, encodedHash string) (bool, error) {
	if password == "" || encodedHash == "" {
		return false, errors.New("credentials missing")
	}
	if len(password) > maxPasswordLength {
		return false, ErrPasswordTooLong
	}

	params, err := parseEncodedHash(encodedHash)
	if err != nil {
		return false, err
	}

	if err := h.acquire(ctx); err != nil {
		return false, err
	}
	defer h.release()

	expectHash := argon2.IDKey(
		[]byte(password), params.salt, params.time, params.memory, params.threads, uint32(len(params.hash)),
	)
	return subtle.ConstantTimeCompare(expectHash, params.hash) == 1, nil
}

func (h *hasher) needsRehash(encodedHash string) bool {
	params, err := parseEncodedHash(encodedHash)
	if err != nil {
		return true
	}
	return params.memory != h.config.Memory ||
		params.time != h.config.Time ||
		params.threads != h.config.Threads ||
		uint32(len(params.hash)) != h.config.KeyLength
}

type hashParams struct {
	memory  uint32
	time    uint32
	threads uint8
	salt    []byte
	hash    []byte
}

func parseEncodedHash(encodedHash string) (*hashParams, error) {
	p := strings.Split(encodedHash, "$")
	if len(p) != 6 || p[1] != defaultVariant {
		return nil, ErrInvalidHash
	}

	if !strings.HasPrefix(p[2], "v=") {
		return nil, ErrInvalidHash
	}
	v, err := strconv.Atoi(p[2][2:])
	if err != nil || v != defaultVersion {
		return nil, ErrInvalidHash
	}

	params := strings.Split(p[3], ",")
	if len(params) != 3 {
		return nil, ErrInvalidHash
	}

	var m, t uint32
	var th uint8
	for _, param := range params {
		kv := strings.Split(param, "=")
		if len(kv) != 2 {
			return nil, ErrInvalidHash
		}
		val, err := strconv.ParseUint(kv[1], 10, 32)
		if err != nil {
			return nil, ErrInvalidHash
		}
		switch kv[0] {
		case "m":
			m = uint32(val)
		case "t":
			t = uint32(val)
		case "p":
			if val > 255 {
				return nil, ErrInvalidHash
			}
			th = uint8(val)
		default:
			return nil, ErrInvalidHash
		}
	}

	if m == 0 || t == 0 || th == 0 {
		return nil, ErrInvalidHash
	}
	if m > maxDecodedMemory || t > maxDecodedTime || th > maxDecodedThreads {
		return nil, errors.New("cryptographic parameters exceed allowed bounds")
	}

	salt, err := base64.RawStdEncoding.DecodeString(p[4])
	if err != nil {
		return nil, ErrInvalidHash
	}
	actualHash, err := base64.RawStdEncoding.DecodeString(p[5])
	if err != nil {
		return nil, ErrInvalidHash
	}
	if len(actualHash) < 16 || len(actualHash) > 128 {
		return nil, ErrInvalidHash
	}

	return &hashParams{memory: m, time: t, threads: th, salt: salt, hash: actualHash}, nil
}

func (h *hasher) acquire(ctx context.Context) error {
	select {
	case h.semaphore <- struct{}{}:
		return nil
	default:
	}

	timer := time.NewTimer(h.config.WaitTimeout)
	defer timer.Stop()

	select {
	case h.semaphore <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return ErrBusy
	}
}

func (h *hasher) release() { <-h.semaphore }
