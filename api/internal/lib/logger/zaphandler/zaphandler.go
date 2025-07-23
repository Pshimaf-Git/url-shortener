package zaphandler

import (
	"context"
	"log/slog"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ slog.Handler = &zapHandler{}

const (
	dev   = "dev"
	prod  = "prod"
	local = "local"
)

type zapHandler struct {
	logger *zap.Logger
}

func (h *zapHandler) Sync() error {
	return h.logger.Sync()
}

func NewProduction(lvl slog.Level, options ...zap.Option) (*zapHandler, error) {
	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(ConvertLevel(lvl)),
		Development:       false,
		DisableCaller:     true,
		Sampling:          nil,
		Encoding:          "json",
		EncoderConfig:     zap.NewProductionEncoderConfig(),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: false,
	}

	logger, err := zapCfg.Build(options...)
	return &zapHandler{logger: logger}, err
}

func NewDevelopment(lvl slog.Level, options ...zap.Option) (*zapHandler, error) {
	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(ConvertLevel(lvl)),
		Development:       true,
		DisableCaller:     true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: false,
	}

	logger, err := zapCfg.Build(options...)
	return &zapHandler{logger: logger}, err
}

func NewLocal(lvl slog.Level, options ...zap.Option) (*zapHandler, error) {
	zapCfg := zap.Config{
		Level:             zap.NewAtomicLevelAt(ConvertLevel(lvl)),
		Development:       true,
		DisableCaller:     true,
		Sampling:          nil,
		Encoding:          "console",
		EncoderConfig:     zap.NewDevelopmentEncoderConfig(),
		OutputPaths:       []string{"stderr"},
		ErrorOutputPaths:  []string{"stderr"},
		DisableStacktrace: false,
	}

	zapCfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := zapCfg.Build(options...)
	return &zapHandler{logger: logger}, err
}

func New(env string, lvl slog.Level, options ...zap.Option) (*zapHandler, error) {
	const fn = "lib.zap.NewZapHandler"
	var (
		logger *zapHandler
		err    error
	)

	switch strings.ToLower(strings.TrimSpace(env)) {
	case prod:
		logger, err = NewProduction(lvl, options...)
	case dev:
		logger, err = NewDevelopment(lvl, options...)
	case local:
		logger, err = NewLocal(lvl, options...)
	default:
		logger, err = NewProduction(lvl, options...)
		slog.Warn("unknown env, using production logger", slog.String("fn", fn), slog.String("env", env))
	}

	return logger, err
}

func NewLogger(env string, lvl slog.Level, options ...zap.Option) (*slog.Logger, func() error, error) {
	handler, err := New(env, lvl, options...)
	if err != nil {
		return nil, nil, err
	}
	return slog.New(handler), handler.Sync, nil
}

func (h *zapHandler) Logger() *zap.Logger {
	return h.logger
}

func (h *zapHandler) Enabled(_ context.Context, level slog.Level) bool {
	var zapLevel zapcore.Level
	switch level {
	case slog.LevelDebug:
		zapLevel = zapcore.DebugLevel
	case slog.LevelInfo:
		zapLevel = zapcore.InfoLevel
	case slog.LevelWarn:
		zapLevel = zapcore.WarnLevel
	case slog.LevelError:
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}
	return h.logger.Core().Enabled(zapLevel)
}

func (h *zapHandler) Handle(_ context.Context, record slog.Record) error {
	fields := make([]zap.Field, 0, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, zap.Any(attr.Key, attr.Value.Any()))
		return true
	})

	switch record.Level {
	case slog.LevelDebug:
		h.logger.Debug(record.Message, fields...)
	case slog.LevelInfo:
		h.logger.Info(record.Message, fields...)
	case slog.LevelWarn:
		h.logger.Warn(record.Message, fields...)
	case slog.LevelError:
		h.logger.Error(record.Message, fields...)
	default:
		h.logger.Info(record.Message, fields...)
	}
	return nil
}

func (h *zapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	fields := make([]zap.Field, len(attrs))
	for i, attr := range attrs {
		fields[i] = zap.Any(attr.Key, attr.Value.Any())
	}
	return &zapHandler{logger: h.logger.With(fields...)}
}

func (h *zapHandler) WithGroup(name string) slog.Handler {
	return &zapHandler{logger: h.logger.Named(name)}
}

func ConvertLevel(slvl slog.Level) zapcore.Level {
	var zapLevel zapcore.Level
	switch slvl {
	case slog.LevelDebug:
		zapLevel = zapcore.DebugLevel
	case slog.LevelInfo:
		zapLevel = zapcore.InfoLevel
	case slog.LevelWarn:
		zapLevel = zapcore.WarnLevel
	case slog.LevelError:
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}
	return zapLevel
}
