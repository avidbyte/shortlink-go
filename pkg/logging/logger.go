package logging

import (
	"fmt"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"path/filepath"
	"time"
)

var (
	Logger      *zap.Logger     // 全局 Logger 实例
	AtomicLevel zap.AtomicLevel // 全局共享日志级别
)

func InitLoggerFromConfig() {
	// 从 viper 获取日志配置
	logLevel := viper.GetString("log.level")
	logPath := viper.GetString("log.path")
	logMaxSize := viper.GetInt("log.max_size")
	logMaxBackups := viper.GetInt("log.max_backups")
	logMaxAge := viper.GetInt("log.max_age")
	logCompress := viper.GetBool("log.compress")

	// 设置默认值
	if logLevel == "" {
		logLevel = "info"
	}
	if logPath == "" {
		logPath = "logs/shortlink.log"
	}
	if logMaxSize <= 0 {
		logMaxSize = 10 // MB
	}
	if logMaxBackups <= 0 {
		logMaxBackups = 5
	}
	if logMaxAge <= 0 {
		logMaxAge = 7 // 天
	}

	// 解析日志级别（安全处理无效值）
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		level = zap.InfoLevel // 默认级别
	}

	// 初始化 atomicLevel（用于共享日志级别）
	AtomicLevel = zap.NewAtomicLevelAt(level)

	// 创建编码器配置
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "ts",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
		// 自定义时间格式
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("2006/01/02 - 15:04:05"))
		},
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建核心组件（控制台 + lumberjack 文件）
	var cores []zapcore.Core

	// 控制台输出
	consoleCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(os.Stdout),
		AtomicLevel,
	)

	cores = append(cores, consoleCore)

	// 文件输出（使用 lumberjack）
	// 确保日志目录存在
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		// 修改这里来处理 fmt.Fprintf 的错误
		if _, writeErr := fmt.Fprintf(os.Stderr, "Failed to create log directory: %v\n", err); writeErr != nil {
			// 这里可以根据需要处理写入错误的情况，例如打印错误详情或退出程序
			fmt.Printf("Failed to write error message to stderr: %v\n", writeErr)
			os.Exit(1)
		}
		return
	}

	// 初始化 lumberjack 日志轮转器
	lumberjackLogger := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    logMaxSize,    // 单位：MB
		MaxBackups: logMaxBackups, // 保留多少个备份文件
		MaxAge:     logMaxAge,     // 保留多少天
		Compress:   logCompress,   // 是否压缩旧日志
		LocalTime:  true,          // 使用本地时间
	}

	// 文件核心（JSON 编码器）
	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(lumberjackLogger),
		level,
	)
	cores = append(cores, fileCore)

	// 合并多个 Core
	core := zapcore.NewTee(cores...)

	// 构建 zap.Logger
	Logger = zap.New(core,
		zap.AddCaller(),
		//zap.AddStacktrace(zap.ErrorLevel),
	)

	// 替换全局 logger
	zap.ReplaceGlobals(Logger)

	Logger.Info("InitLoggerFromConfig finished")
}
