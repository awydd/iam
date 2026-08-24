package conf

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

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

type Config struct {
	Env    iamEnv `yaml:"env"`
	Logger Logger `yaml:"logger"`
	HTTP   HTTP   `yaml:"http"`
}

func (c *Config) IsDev() bool {
	return c.Env == envDev
}

func defaultConfig() *Config {
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
