package service

import (
	"os"
	"path/filepath"
	"shortlink-go/internal/i18n"
	"shortlink-go/internal/repository"
	"shortlink-go/pkg/logging"
	"testing"

	"github.com/spf13/viper"
)

func initTestEnv(t *testing.T) {
	root := getProjectRoot()
	if root == "" {
		t.Fatal("cannot find project root")
	}
	// 加载配置文件（与 main.go 中保持一致）
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(root)
	if err := viper.ReadInConfig(); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	logging.InitLoggerFromConfig()

	// 初始化数据库
	repository.InitDB(logging.Logger, logging.AtomicLevel)

	// 初始化 Redis
	repository.InitRedis()

	// 初始化 i18n（可选，根据是否影响逻辑）
	_, err := i18n.InitI18n([]string{
		"../../i18n/en.toml",
		"../../i18n/zh.toml",
	}, "en")
	if err != nil {
		t.Fatalf("Failed to initialize i18n: %v", err)
	}
}

func getProjectRoot() string {
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func TestStatisticalData(t *testing.T) {
	initTestEnv(t)

	if err := StatisticalData(); err != nil {
		t.Errorf("StatisticalData failed: %v", err)
	}
}
