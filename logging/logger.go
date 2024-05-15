package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(logfile *os.File) (*zap.Logger, error) {
	encoderConf := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "name",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileEncoder := zapcore.NewJSONEncoder(encoderConf)
	consoleEncoder := zapcore.NewJSONEncoder(encoderConf)

	core := zapcore.NewTee(
		zapcore.NewCore(fileEncoder, zapcore.AddSync(logfile), zap.InfoLevel),
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel),
	)
	l := zap.New(core)
	return l, nil
}
