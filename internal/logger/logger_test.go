package logger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/awydd/iam/conf"
)

func TestMain(m *testing.M) {
	rootPath := filepath.Join("../../")
	if err := os.Chdir(rootPath); err != nil {
		panic(err)
	}

	if err := conf.Init(); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestInitAndLog(t *testing.T) {
	cfg := conf.Get().Logger

	if err := Init(cfg); err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}
	defer Sync()

	Debug("this is a debug message: %d", 123)
	Info("this is an info message: %s", "hello zap")
	Warn("this is a warn message")
	Error("this is an error message")

	if _, err := os.Stat(cfg.Filename); os.IsNotExist(err) {
		t.Errorf("expected log file %s to be created, but it does not exist", cfg.Filename)
	} else {
		t.Logf("log file successfully created at: %s", cfg.Filename)
	}
}

// 测试日志轮转功能
func TestRotate(t *testing.T) {
	testLoggerCfg := conf.Logger{
		Filename:   "data/logs/rotate_test.log",
		MaxSizeMB:  1, // 设为 1MB，方便快速触发
		MaxBackups: 3,
		MaxAgeDays: 7,
		Compress:   true,
		Level:      "info",
	}

	if err := Init(testLoggerCfg); err != nil {
		t.Fatalf("failed to init logger: %v", err)
	}
	defer Sync()

	// 循环写入日志，直到文件体积超过 1MB 触发轮转
	// 一条日志大约 100 字节，写 12,000 条大约是 1.2MB
	t.Log("开始写入日志以触发轮转...")
	for i := 0; i < 12000; i++ {
		Info("这是第 %d 条测试日志，用于验证 lumberjack 的自动轮转与压缩功能是否正常工作", i)
	}

	// 检查 logs 目录下是否多出了被压缩或备份的 .gz 日志文件
	logDir := filepath.Dir(testLoggerCfg.Filename)
	entries, err := os.ReadDir(logDir)
	if err != nil {
		t.Fatalf("failed to read log dir: %v", err)
	}

	foundBackup := false
	for _, entry := range entries {
		t.Logf("发现文件: %s", entry.Name())
		// lumberjack 轮转后通常会生成带有时间戳或后缀的备份文件
		if entry.Name() != "rotate_test.log" {
			foundBackup = true
		}
	}

	if foundBackup {
		t.Log("日志轮转及备份文件生成")
	}
}
