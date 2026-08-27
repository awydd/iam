package conf

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// 这里的 TestMain 会在测试启动时拦截，先执行 os.Chdir("..")，就不会在当前而是上层目录创建 data
func TestMain(m *testing.M) {
	rootPath := filepath.Join("..")
	if err := os.Chdir(rootPath); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func TestConf(t *testing.T) {
	Init()
	fmt.Println(Get())
}
