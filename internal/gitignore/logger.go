package gitignore

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cicbyte/gig/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// 全局日志实例
	globalLogger *zap.SugaredLogger
	once         sync.Once
)

// LogLevel 日志级别
type LogLevel string

const (
	// LogDebug 调试级别
	LogDebug LogLevel = "debug"
	// LogInfo 信息级别
	LogInfo LogLevel = "info"
	// LogWarn 警告级别
	LogWarn LogLevel = "warn"
	// LogError 错误级别
	LogError LogLevel = "error"
)

// InitLogger 初始化日志系统
func InitLogger(level LogLevel) (*zap.SugaredLogger, error) {
	var err error
	once.Do(func() {
		// 获取日志目录
		logDir, e := getLogDirectory()
		if e != nil {
			err = fmt.Errorf("获取日志目录失败: %w", e)
			return
		}

		// 确保日志目录存在
		if e := os.MkdirAll(logDir, 0755); e != nil {
			err = fmt.Errorf("创建日志目录失败: %w", e)
			return
		}

		// 设置日志文件路径
		logFile := filepath.Join(logDir, fmt.Sprintf("gig-%s.log", time.Now().Format("2006-01-02")))

		// 创建编码器配置
		encoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}

		// 创建日志级别
		var zapLevel zapcore.Level
		switch level {
		case LogDebug:
			zapLevel = zapcore.DebugLevel
		case LogInfo:
			zapLevel = zapcore.InfoLevel
		case LogWarn:
			zapLevel = zapcore.WarnLevel
		case LogError:
			zapLevel = zapcore.ErrorLevel
		default:
			zapLevel = zapcore.InfoLevel
		}

		// 创建Core
		fileWriter, e := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if e != nil {
			err = fmt.Errorf("打开日志文件失败: %w", e)
			return
		}

		core := zapcore.NewTee(
			zapcore.NewCore(
				zapcore.NewJSONEncoder(encoderConfig),
				zapcore.AddSync(fileWriter),
				zapLevel,
			),
			zapcore.NewCore(
				zapcore.NewConsoleEncoder(encoderConfig),
				zapcore.AddSync(os.Stderr),
				zapLevel,
			),
		)

		// 创建Logger
		logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
		globalLogger = logger.Sugar()
	})

	if err != nil {
		// 如果初始化失败，创建一个基本的控制台日志器
		config := zap.NewDevelopmentConfig()
		logger, _ := config.Build()
		globalLogger = logger.Sugar()
		return globalLogger, err
	}

	return globalLogger, nil
}

// GetLogger 获取全局日志实例
func GetLogger() *zap.SugaredLogger {
	if globalLogger == nil {
		// 如果尚未初始化，使用默认配置初始化
		InitLogger(LogInfo)
	}
	return globalLogger
}

// getLogDirectory 获取日志目录
func getLogDirectory() (string, error) {
	home, err := utils.GetUserHomeDir()
	if err != nil {
		return "", err
	}

	// 与配置文件同级目录
	return filepath.Join(home, ".cicbyte", "gig", "logs"), nil
}

// Debug 记录调试级别日志
func Debug(format string, args ...any) {
	GetLogger().Debugf(format, args...)
}

// Info 记录信息级别日志
func Info(format string, args ...any) {
	GetLogger().Infof(format, args...)
}

// Warn 记录警告级别日志
func Warn(format string, args ...any) {
	GetLogger().Warnf(format, args...)
}

// Error 记录错误级别日志
func Error(format string, args ...any) {
	GetLogger().Errorf(format, args...)
}

// Fatal 记录致命错误并退出
func Fatal(format string, args ...any) {
	GetLogger().Fatalf(format, args...)
}

// WithFields 创建带有字段的日志记录器
func WithFields(fields map[string]any) *zap.SugaredLogger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return GetLogger().With(args...)
}
