package pkg

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

type Logger struct {
	Logger *zap.Logger
}

func NewLogger() *Logger {
	encodeConfig := zap.NewProductionConfig()
	encodeConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encodeConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encodeConfig.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	file, _ := os.OpenFile("./log/log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	syncFile := zapcore.AddSync(file)
	syncConsole := zapcore.AddSync(os.Stderr)

	core := zapcore.NewCore(zapcore.NewJSONEncoder(encodeConfig.EncoderConfig), zapcore.NewMultiWriteSyncer(
		syncFile, syncConsole), zapcore.InfoLevel)
	logger := zap.New(core, zap.AddCaller())

	return &Logger{Logger: logger}
}
