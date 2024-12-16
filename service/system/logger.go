package system

import (
	"fmt"
	"os"

	"github.com/shipeng101/singdns/pkg/types"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var logger *zap.Logger

// InitLogger 初始化日志
func InitLogger(config types.LogConfig) error {
	// 创建日志目录
	if err := os.MkdirAll("logs", 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 设置日志级别
	var level zapcore.Level
	switch config.Level {
	case "debug":
		level = zap.DebugLevel
	case "info":
		level = zap.InfoLevel
	case "warn":
		level = zap.WarnLevel
	case "error":
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	// 配置编码器
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// 配置输出
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// 配置日志轮转
	writer := &lumberjack.Logger{
		Filename:   config.File,
		MaxSize:    config.MaxSize,
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,
		Compress:   config.Compress,
	}

	// 创建核心
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level),
		zapcore.NewCore(fileEncoder, zapcore.AddSync(writer), level),
	)

	// 创建日志记录器
	logger = zap.New(core)

	return nil
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Debug(msg, fields...)
	}
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Info(msg, fields...)
	}
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Warn(msg, fields...)
	}
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Error(msg, fields...)
	}
}

// Fatal 记录致命错误日志
func Fatal(msg string, fields ...zap.Field) {
	if logger != nil {
		logger.Fatal(msg, fields...)
	}
}

// Sync 同步日志
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}
