package conf

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/awydd/iam/pkg/random"
	"github.com/goccy/go-yaml"
)

type iamEnv string

const (
	envDev  iamEnv = "dev"
	envProd iamEnv = "prod"
)

func (e iamEnv) valid() bool {
	switch e {
	case envDev, envProd:
		return true
	default:
		return false
	}
}

type Logger struct {
	Filename   string `yaml:"filename"`
	MaxSizeMB  int    `yaml:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAgeDays int    `yaml:"max_age_days"`
	Compress   bool   `yaml:"compress"`
	Level      string `yaml:"level"`
}

type HTTP struct {
	Port                 string        `yaml:"port"`
	MaxConnections       int           `yaml:"max_connections"`
	ReadTimeout          time.Duration `yaml:"read_timeout"`
	ReadHeaderTimeout    time.Duration `yaml:"read_header_timeout"`
	WriteTimeout         time.Duration `yaml:"write_timeout"`
	IdleTimeout          time.Duration `yaml:"idle_timeout"`
	MaxHeaderBytes       int           `yaml:"max_header_bytes"`
	ShutdownTimeout      time.Duration `yaml:"shutdown_timeout"`
	TrustedProxies       []string      `yaml:"trusted_proxies"`
	EnablePProf          bool          `yaml:"enable_pprof"`
	BlockProfileRate     int           `yaml:"block_profile_rate"`
	MutexProfileFraction int           `yaml:"mutex_profile_fraction"`
	CORSAllowOrigins     []string      `yaml:"cors_allow_origins"`
}

type Redis struct {
	Addr            string        `yaml:"addr"`
	Password        string        `yaml:"password"`
	DB              int           `yaml:"db"`
	Prefix          string        `yaml:"prefix"`
	PoolSize        int           `yaml:"pool_size"`
	MinIdleConns    int           `yaml:"min_idle_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	MaxRetries      int           `yaml:"max_retries"`
	DialTimeout     time.Duration `yaml:"dial_timeout"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	PoolTimeout     time.Duration `yaml:"pool_timeout"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

type Database struct {
	Host         string        `yaml:"host"`
	Port         string        `yaml:"port"`
	User         string        `yaml:"user"`
	Password     string        `yaml:"password"`
	DBName       string        `yaml:"db_name"`
	Timezone     string        `yaml:"timezone"`
	MaxIdleConns int           `yaml:"max_idle_conns"`
	MaxOpenConns int           `yaml:"max_open_conns"`
	MaxLifetime  time.Duration `yaml:"max_lifetime"`
	MaxIdleTime  time.Duration `yaml:"max_idle_time"`
}

func (d *Database) DSN() string {
	tz := d.Timezone
	if tz == "" {
		tz = "Local"
	}

	params := url.Values{}
	params.Set("charset", "utf8mb4")
	params.Set("parseTime", "True")
	params.Set("loc", tz)

	return fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?%s",
		d.User,
		d.Password,
		d.Host,
		d.Port,
		d.DBName,
		params.Encode(),
	)
}

type Password struct {
	Time           uint32        `yaml:"time"`
	Memory         uint32        `yaml:"memory"`
	Threads        uint8         `yaml:"threads"`
	KeyLength      uint32        `yaml:"key_length"`
	SaltLen        uint32        `yaml:"salt_len"`
	MaxConcurrency int           `yaml:"max_concurrency"`
	WaitTimeout    time.Duration `yaml:"wait_timeout"`
}

type JWT struct {
	Issuer          string        `yaml:"issuer"`
	Audience        string        `yaml:"audience"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl"`
}

type Cookie struct {
	Domain   string `yaml:"domain"`
	Path     string `yaml:"path"`
	Secure   bool   `yaml:"secure"`
	HttpOnly bool   `yaml:"http_only"`
	SameSite string `yaml:"same_site"` // lax, strict, none
	Secret   string `yaml:"secret"`
}

const (
	DefaultKeypairVerifyCacheTTL  = 5 * time.Minute
	DefaultKeypairSigningCacheTTL = 10 * time.Minute
)

type Keypair struct {
	VerifyCacheTTL  time.Duration `yaml:"verify_cache_ttl"`
	SigningCacheTTL time.Duration `yaml:"signing_cache_ttl"`
}

type Config struct {
	Env      iamEnv   `yaml:"env"`
	Logger   Logger   `yaml:"logger"`
	HTTP     HTTP     `yaml:"http"`
	Redis    Redis    `yaml:"redis"`
	Database Database `yaml:"database"`
	Password Password `yaml:"password"`
	JWT      JWT      `yaml:"jwt"`
	Cookie   Cookie   `yaml:"cookie"`
	Keypair  Keypair  `yaml:"keypair"`
}

func (c *Config) IsDev() bool {
	return c.Env == envDev
}

func defaultConfig() *Config {
	mustGenKey := func() string {
		if s, err := random.Alphanumeric(32); err == nil {
			return s
		}
		// openssl rand -base64 48 | tr -dc 'a-zA-Z0-9' | head -c 32
		// head -c 32 /dev/urandom | base64 | tr -dc 'a-zA-Z0-9' | head -c 32
		return "B3iNF9vrTob6YVsbp8wP9BgXfU3WkO3D"
	}

	return &Config{
		Env: envProd,
		Logger: Logger{
			Filename:   filepath.Join(DataDir, "logs/iam.log"),
			MaxSizeMB:  100,
			MaxBackups: 7,
			MaxAgeDays: 30,
			Compress:   true,
			// debug < info < warn < error < fatal
			// info 时会记录到 info & warn & error & fatal
			Level: "info",
		},
		HTTP: HTTP{
			Port:              "26824",
			MaxConnections:    0,
			ReadTimeout:       5 * time.Second,
			ReadHeaderTimeout: 5 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       60 * time.Second,
			MaxHeaderBytes:    1 << 20, // 1MB
			ShutdownTimeout:   10 * time.Second,
			TrustedProxies: []string{
				"127.0.0.0/8",    // IPv4 Loopback
				"::1/128",        // IPv6 Loopback
				"10.0.0.0/8",     // Private Network
				"172.16.0.0/12",  // Private Network / Docker Default Bridge
				"192.168.0.0/16", // Private Network
				"fc00::/7",       // Unique Local Address (IPv6)
			},
			EnablePProf:          false,
			BlockProfileRate:     0,
			MutexProfileFraction: 0,
			CORSAllowOrigins:     nil,
		},
		Redis: Redis{
			Addr:            "127.0.0.1:6379",
			Password:        "",
			DB:              0,
			Prefix:          "iam",
			PoolSize:        100,
			MinIdleConns:    10,
			MaxIdleConns:    50,
			MaxRetries:      3,
			DialTimeout:     5 * time.Second,
			ReadTimeout:     3 * time.Second,
			WriteTimeout:    3 * time.Second,
			PoolTimeout:     4 * time.Second,
			ConnMaxIdleTime: 30 * time.Minute,
			ConnMaxLifetime: time.Hour,
		},
		Database: Database{
			Host:         "127.0.0.1",
			Port:         "3306",
			User:         "root",
			Password:     "",
			DBName:       "iam",
			Timezone:     "Local",
			MaxIdleConns: 10,
			MaxOpenConns: 100,
			MaxLifetime:  1 * time.Hour,
			MaxIdleTime:  5 * time.Minute,
		},
		Password: Password{
			Time:           3,
			Memory:         32 * 1024, // 32MB
			Threads:        2,
			KeyLength:      32,
			SaltLen:        16,
			MaxConcurrency: 8,
			WaitTimeout:    5 * time.Second,
		},
		JWT: JWT{
			Issuer:          "iam",
			Audience:        "",
			AccessTokenTTL:  2 * time.Hour,
			RefreshTokenTTL: 7 * 24 * time.Hour,
		},
		Cookie: Cookie{
			Domain:   "",
			Path:     "/",
			Secure:   true,
			HttpOnly: true,
			SameSite: "lax",
			Secret:   mustGenKey(),
		},
		Keypair: Keypair{
			VerifyCacheTTL:  DefaultKeypairVerifyCacheTTL,
			SigningCacheTTL: DefaultKeypairSigningCacheTTL,
		},
	}
}

var (
	once    sync.Once
	initErr error

	mu      sync.RWMutex
	current *Config
)

const (
	// 数据目录（配置文件、缓存等）
	DataDir = "data"
	// 配置文件名
	cfgFilename = "config.yaml"
)

func cfgPath() string {
	return filepath.Join(DataDir, cfgFilename)
}

func ensure() error {
	p := cfgPath()
	_, err := os.Stat(p)
	if err == nil {
		return nil
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("conf: check file status: %w", err)
	}

	dir := filepath.Dir(p)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("conf: create config directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(defaultConfig())
	if err != nil {
		return fmt.Errorf("conf: marshal default config: %w", err)
	}

	if err := os.WriteFile(p, data, 0644); err != nil {
		return fmt.Errorf("conf: write default config file %s: %w", p, err)
	}

	return nil
}

func validate(cfg *Config) error {
	var errs []error

	// 环境
	if !cfg.Env.valid() {
		errs = append(errs, fmt.Errorf("env %q is invalid (allowed: dev, prod)", cfg.Env))
	}

	// 日志
	if cfg.Logger.Filename == "" {
		errs = append(errs, errors.New("logger.filename must not be empty"))
	}
	if cfg.Logger.MaxSizeMB <= 0 {
		errs = append(errs, errors.New("logger.max_size_mb must be positive"))
	}
	if cfg.Logger.MaxBackups < 0 {
		errs = append(errs, errors.New("logger.max_backups cannot be negative"))
	}
	if cfg.Logger.MaxAgeDays < 0 {
		errs = append(errs, errors.New("logger.max_age_days cannot be negative"))
	}
	switch cfg.Logger.Level {
	case "debug", "info", "warn", "error":
		// valid levels
	default:
		errs = append(errs, fmt.Errorf("logger.level %q is invalid (allowed: debug, info, warn, error)", cfg.Logger.Level))
	}

	// HTTP
	if cfg.HTTP.Port == "" {
		errs = append(errs, errors.New("http.port must not be empty"))
	}
	if cfg.HTTP.MaxConnections < 0 {
		errs = append(errs, errors.New("http.max_connections cannot be negative"))
	}
	if cfg.HTTP.ReadTimeout <= 0 {
		errs = append(errs, errors.New("http.read_timeout must be positive"))
	}
	if cfg.HTTP.ReadHeaderTimeout <= 0 {
		errs = append(errs, errors.New("http.read_header_timeout must be positive"))
	}
	if cfg.HTTP.WriteTimeout <= 0 {
		errs = append(errs, errors.New("http.write_timeout must be positive"))
	}
	if cfg.HTTP.IdleTimeout <= 0 {
		errs = append(errs, errors.New("http.idle_timeout must be positive"))
	}
	if cfg.HTTP.ShutdownTimeout <= 0 {
		errs = append(errs, errors.New("http.shutdown_timeout must be positive"))
	}
	if cfg.HTTP.MaxHeaderBytes <= 0 {
		errs = append(errs, errors.New("http.max_header_bytes must be positive"))
	}
	if cfg.HTTP.BlockProfileRate < 0 {
		errs = append(errs, errors.New("http.block_profile_rate cannot be negative"))
	}
	if cfg.HTTP.MutexProfileFraction < 0 {
		errs = append(errs, errors.New("http.mutex_profile_fraction cannot be negative"))
	}
	// TrustedProxies validation
	if len(cfg.HTTP.TrustedProxies) == 0 {
		errs = append(errs, errors.New("http.trusted_proxies cannot be empty"))
	} else {
		for i, cidr := range cfg.HTTP.TrustedProxies {
			c := strings.TrimSpace(cidr)
			if c == "" {
				errs = append(errs, fmt.Errorf("http.trusted_proxies[%d] cannot be empty", i))
				continue
			}

			if !strings.Contains(c, "/") {
				if ip := net.ParseIP(c); ip == nil {
					errs = append(errs, fmt.Errorf("http.trusted_proxies[%d] %q is not a valid IP or CIDR block", i, cidr))
				}
			} else {
				if _, _, err := net.ParseCIDR(c); err != nil {
					errs = append(errs, fmt.Errorf("http.trusted_proxies[%d] %q is not a valid CIDR block: %v", i, cidr, err))
				}
			}
		}
	}

	// Redis
	if cfg.Redis.Addr == "" {
		errs = append(errs, errors.New("redis.addr must not be empty"))
	}
	if cfg.Redis.DB < 0 {
		errs = append(errs, errors.New("redis.db cannot be negative"))
	}
	if cfg.Redis.PoolSize <= 0 {
		errs = append(errs, errors.New("redis.pool_size must be greater than 0"))
	}
	if cfg.Redis.MinIdleConns < 0 {
		errs = append(errs, errors.New("redis.min_idle_conns cannot be negative"))
	}
	if cfg.Redis.MinIdleConns > cfg.Redis.PoolSize {
		errs = append(errs, errors.New("redis.min_idle_conns cannot exceed pool_size"))
	}
	if cfg.Redis.MaxIdleConns < 0 {
		errs = append(errs, errors.New("redis.max_idle_conns cannot be negative"))
	}
	if cfg.Redis.MaxIdleConns > 0 && cfg.Redis.MaxIdleConns > cfg.Redis.PoolSize {
		errs = append(errs, errors.New("redis.max_idle_conns cannot exceed pool_size"))
	}
	if cfg.Redis.MaxIdleConns > 0 && cfg.Redis.MinIdleConns > cfg.Redis.MaxIdleConns {
		errs = append(errs, errors.New("redis.min_idle_conns cannot exceed max_idle_conns"))
	}
	if cfg.Redis.MaxRetries < 0 {
		errs = append(errs, errors.New("redis.max_retries cannot be negative"))
	}
	if cfg.Redis.DialTimeout <= 0 {
		errs = append(errs, errors.New("redis.dial_timeout must be positive"))
	}
	if cfg.Redis.ReadTimeout <= 0 {
		errs = append(errs, errors.New("redis.read_timeout must be positive"))
	}
	if cfg.Redis.WriteTimeout <= 0 {
		errs = append(errs, errors.New("redis.write_timeout must be positive"))
	}
	if cfg.Redis.PoolTimeout <= 0 {
		errs = append(errs, errors.New("redis.pool_timeout must be positive"))
	}
	if cfg.Redis.ConnMaxIdleTime < 0 {
		errs = append(errs, errors.New("redis.conn_max_idle_time cannot be negative"))
	}
	if cfg.Redis.ConnMaxLifetime < 0 {
		errs = append(errs, errors.New("redis.conn_max_lifetime cannot be negative"))
	}
	if cfg.Redis.ConnMaxLifetime > 0 && cfg.Redis.ConnMaxIdleTime > cfg.Redis.ConnMaxLifetime {
		errs = append(errs, errors.New("redis.conn_max_idle_time should not exceed conn_max_lifetime"))
	}

	// 数据库
	if cfg.Database.Host == "" {
		errs = append(errs, errors.New("database.host must not be empty"))
	}
	if cfg.Database.Port == "" {
		errs = append(errs, errors.New("database.port must not be empty"))
	}
	if cfg.Database.User == "" {
		errs = append(errs, errors.New("database.user must not be empty"))
	}
	if cfg.Database.DBName == "" {
		errs = append(errs, errors.New("database.db_name must not be empty"))
	}
	if cfg.Database.MaxIdleConns < 0 {
		errs = append(errs, errors.New("database.max_idle_conns cannot be negative"))
	}
	if cfg.Database.MaxOpenConns <= 0 {
		errs = append(errs, errors.New("database.max_open_conns must be greater than 0"))
	}
	if cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		errs = append(errs, errors.New("database.max_idle_conns cannot exceed max_open_conns"))
	}
	if cfg.Database.MaxLifetime < 0 {
		errs = append(errs, errors.New("database.max_lifetime cannot be negative"))
	}
	if cfg.Database.MaxIdleTime < 0 {
		errs = append(errs, errors.New("database.max_idle_time cannot be negative"))
	}

	// 密码
	if cfg.Password.Time == 0 {
		errs = append(errs, errors.New("password.time must be greater than 0"))
	}
	if cfg.Password.Memory == 0 {
		errs = append(errs, errors.New("password.memory must be greater than 0"))
	}
	if cfg.Password.Threads == 0 {
		errs = append(errs, errors.New("password.threads must be greater than 0"))
	}
	if cfg.Password.KeyLength == 0 {
		errs = append(errs, errors.New("password.key_length must be greater than 0"))
	}
	if cfg.Password.SaltLen == 0 {
		errs = append(errs, errors.New("password.salt_len must be greater than 0"))
	}
	if cfg.Password.MaxConcurrency <= 0 {
		errs = append(errs, errors.New("password.max_concurrency must be greater than 0"))
	}
	if cfg.Password.WaitTimeout <= 0 {
		errs = append(errs, errors.New("password.wait_timeout must be positive"))
	}

	// 密钥
	if cfg.Keypair.VerifyCacheTTL < 0 {
		errs = append(errs, errors.New("keypair.verify_cache_ttl must not be negative"))
	}
	if cfg.Keypair.SigningCacheTTL < 0 {
		errs = append(errs, errors.New("keypair.signing_cache_ttl must not be negative"))
	}

	return errors.Join(errs...)
}

func load() (*Config, error) {
	if err := ensure(); err != nil {
		return nil, err
	}

	p := cfgPath()
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("conf: read file %s: %w", p, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("conf: parse yaml %s: %w", p, err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("conf: validate: %w", err)
	}

	return &cfg, nil
}

func Init() error {
	once.Do(func() {
		if err := os.MkdirAll(DataDir, 0755); err != nil {
			initErr = fmt.Errorf("conf: create directory %s: %w", DataDir, err)
			return
		}

		cfg, err := load()
		if err != nil {
			initErr = err
			return
		}

		mu.Lock()
		current = cfg
		mu.Unlock()
	})
	return initErr
}

func Get() *Config {
	mu.RLock()
	defer mu.RUnlock()
	if current == nil {
		panic("conf: not initialized, call Init() first")
	}
	return current
}
