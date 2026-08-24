package conf

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

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

type Config struct {
	Env    iamEnv `yaml:"env"`
	Logger Logger `yaml:"logger"`
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
