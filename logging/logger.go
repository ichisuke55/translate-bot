package logging

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger() (*zap.Logger, error) {
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
	// conf := zap.Config{
	// 	Level:         zap.NewAtomicLevel(),
	// 	Development:   false,
	// 	Encoding:      "json",
	// 	EncoderConfig: encoderConf,
	// }
	// logger, err := conf.Build()
	// if err != nil {
	// 	return nil, err
	// }
	consoleEncoder := zapcore.NewJSONEncoder(encoderConf)
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zap.InfoLevel),
	)
	l := zap.New(core)
	return l, nil
}
