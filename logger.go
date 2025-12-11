package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// InitLogger 初始化日志记录器
func InitLogger() error {
	// 创建logs目录
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 获取当前日期
	currentDate := time.Now().Format("2006-01-02")
	logFile := filepath.Join(logsDir, fmt.Sprintf("zabbix-mcp-%s.log", currentDate))

	// 配置日志编码器
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 创建JSON编码器
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// 创建文件写入器
	fileWriter := zapcore.AddSync(&dateRotatingWriter{
		filename: logFile,
		file:     nil,
	})

	// 创建控制台写入器
	consoleWriter := zapcore.Lock(os.Stdout)

	// 创建多写入器（同时写入文件和控制台）
	core := zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, fileWriter, zap.InfoLevel),
		zapcore.NewCore(jsonEncoder, consoleWriter, zap.InfoLevel),
	)

	// 创建logger
	logger = zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	sugar = logger.Sugar()

	return nil
}

// GetLogger 获取logger实例
func GetLogger() *zap.Logger {
	return logger
}

// GetSugar 获取sugar logger实例
func GetSugar() *zap.SugaredLogger {
	return sugar
}

// Sync 同步日志
func Sync() {
	if logger != nil {
		logger.Sync()
	}
}

// dateRotatingWriter 按日期切分的日志写入器
type dateRotatingWriter struct {
	filename string
	file     *os.File
	date     string
}

func (w *dateRotatingWriter) Write(p []byte) (n int, err error) {
	currentDate := time.Now().Format("2006-01-02")

	// 检查是否需要切换日志文件
	if w.date != currentDate || w.file == nil {
		if w.file != nil {
			w.file.Close()
		}

		// 创建新的日志文件
		newFilename := filepath.Join("logs", fmt.Sprintf("zabbix-mcp-%s.log", currentDate))
		w.file, err = os.OpenFile(newFilename, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		w.date = currentDate
		w.filename = newFilename
	}

	return w.file.Write(p)
}

func (w *dateRotatingWriter) Sync() error {
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}
