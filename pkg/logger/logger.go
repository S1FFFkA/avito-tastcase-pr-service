package logger

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func InitLogger() {
	dev := false
	encoderCfg := zap.NewProductionEncoderConfig()
	level := zap.NewAtomicLevelAt(zap.InfoLevel)

	encoderCfg.TimeKey = "timestamp"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	// Создаем директорию logs, если её нет
	if err := os.MkdirAll("logs", 0755); err != nil {
		log.Printf("Warning: failed to create logs directory: %v", err)
	}

	config := zap.Config{
		Level:             level,
		Development:       dev,
		DisableStacktrace: true,
		DisableCaller:     false,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     encoderCfg,
		OutputPaths: []string{
			"logs/log.txt",
			"stdout",
		},
		ErrorOutputPaths: []string{
			"logs/error.txt",
			"stderr",
		},
		InitialFields: map[string]interface{}{
			"pid": os.Getpid(),
		},
	}

	baseLogger, err := config.Build()
	if err != nil {
		log.Fatal("Error building zap logger")
	}

	Logger = baseLogger.Sugar()
}

func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// SafeInfow безопасно вызывает Infow, если логгер инициализирован
func SafeInfow(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Infow(msg, keysAndValues...)
	}
}

// SafeErrorw безопасно вызывает Errorw, если логгер инициализирован
func SafeErrorw(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Errorw(msg, keysAndValues...)
	}
}
