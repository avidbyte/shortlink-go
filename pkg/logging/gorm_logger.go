package logging

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm/logger"
	"time"
)

type gormZapLogger struct {
	logger *zap.Logger
	level  logger.LogLevel
}

func ToGormLogLevel(zapLevel zapcore.Level) logger.LogLevel {
	switch zapLevel {
	case zapcore.DebugLevel, zapcore.InfoLevel:
		return logger.Info
	case zapcore.WarnLevel:
		return logger.Warn
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return logger.Error
	default:
		return logger.Silent
	}
}

func NewGormLogger(l *zap.Logger, level logger.LogLevel) logger.Interface {
	return &gormZapLogger{
		logger: l,
		level:  level,
	}
}

func (g *gormZapLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &gormZapLogger{
		logger: g.logger,
		level:  level,
	}
}

func (g *gormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Info {
		g.logger.Info(fmt.Sprintf(msg, data...))
	}
}

func (g *gormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Warn {
		g.logger.Warn(fmt.Sprintf(msg, data...))
	}
}

func (g *gormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if g.level >= logger.Error {
		g.logger.Error(fmt.Sprintf(msg, data...))
	}
}

func (g *gormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if g.level >= logger.Info {
		elapsed := time.Since(begin)
		sql, rows := fc()
		g.logger.Info("GORM SQL",
			zap.Duration("duration", elapsed),
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Error(err),
		)
	}
}
