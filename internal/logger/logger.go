/*
日志模块

提供结构化日志记录功能 (基于 zap)
*/
package logger

import (
	"siriusec-mcp/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Log 全局日志实例
	Log *zap.Logger
	// Sugar 全局 SugaredLogger 实例
	Sugar *zap.SugaredLogger
)

func init() {
	Log, Sugar = NewLogger()
}

// NewLogger 创建新的日志实例
func NewLogger() (*zap.Logger, *zap.SugaredLogger) {
	cfg := zap.NewProductionConfig()

	// 设置输出到 stderr (避免干扰 MCP 协议的 stdout 通信)
	cfg.OutputPaths = []string{"stderr"}
	cfg.ErrorOutputPaths = []string{"stderr"}

	// 设置日志级别
	level := zapcore.InfoLevel
	if config.GlobalConfig != nil {
		level = parseLogLevel(config.GlobalConfig.Log.Level)
	}
	cfg.Level = zap.NewAtomicLevelAt(level)

	// 设置编码格式
	cfg.Encoding = "json"
	cfg.EncoderConfig = zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	logger, err := cfg.Build(zap.AddCallerSkip(1))
	if err != nil {
		// 回退到基本配置
		logger = zap.NewNop()
	}

	return logger, logger.Sugar()
}

// parseLogLevel 解析日志级别
func parseLogLevel(level string) zapcore.Level {
	switch level {
	case "DEBUG":
		return zapcore.DebugLevel
	case "INFO":
		return zapcore.InfoLevel
	case "WARN", "WARNING":
		return zapcore.WarnLevel
	case "ERROR":
		return zapcore.ErrorLevel
	case "FATAL":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// SetLevel 设置日志级别
func SetLevel(level string) {
	// zap 不支持动态修改级别，需要重新创建 logger
	// 这里仅作兼容性保留
}

// Debug 记录调试日志
func Debug(msg string, fields ...zap.Field) {
	Log.Debug(msg, fields...)
}

// Debugf 记录格式化的调试日志
func Debugf(template string, args ...interface{}) {
	Sugar.Debugf(template, args...)
}

// Info 记录信息日志
func Info(msg string, fields ...zap.Field) {
	Log.Info(msg, fields...)
}

// Infof 记录格式化的信息日志
func Infof(template string, args ...interface{}) {
	Sugar.Infof(template, args...)
}

// Warn 记录警告日志
func Warn(msg string, fields ...zap.Field) {
	Log.Warn(msg, fields...)
}

// Warnf 记录格式化的警告日志
func Warnf(template string, args ...interface{}) {
	Sugar.Warnf(template, args...)
}

// Error 记录错误日志
func Error(msg string, fields ...zap.Field) {
	Log.Error(msg, fields...)
}

// Errorf 记录格式化的错误日志
func Errorf(template string, args ...interface{}) {
	Sugar.Errorf(template, args...)
}

// Fatal 记录致命错误日志并退出
func Fatal(msg string, fields ...zap.Field) {
	Log.Fatal(msg, fields...)
}

// Fatalf 记录格式化的致命错误日志并退出
func Fatalf(template string, args ...interface{}) {
	Sugar.Fatalf(template, args...)
}

// With 创建带字段的日志记录器
func With(fields ...zap.Field) *zap.Logger {
	return Log.With(fields...)
}

// Sync 刷新日志缓冲区
func Sync() error {
	return Log.Sync()
}
