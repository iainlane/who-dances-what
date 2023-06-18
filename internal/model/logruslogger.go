package model

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm/logger"
)

type LogrusLogger struct {
	*logrus.Entry
}

func NewLogrusLogger(logger *logrus.Entry) LogrusLogger {
	entry := logger.WithField("source", "gorm")
	return LogrusLogger{entry}
}

func gormLevelToLogrusLevel(level logger.LogLevel) logrus.Level {
	switch level {
	case logger.Silent:
		return logrus.PanicLevel
	case logger.Error:
		return logrus.ErrorLevel
	case logger.Warn:
		return logrus.WarnLevel
	case logger.Info:
		return logrus.InfoLevel
	default:
		return logrus.WarnLevel
	}
}

func (l LogrusLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := l.Entry.Dup()
	newLogger.Level = gormLevelToLogrusLevel(level)
	return LogrusLogger{newLogger}
}

// Info levels in gorm are used for SQL queries. This is more like a debug
// message IMO.
func (l LogrusLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	l.Entry.WithContext(ctx).Debugf(msg, data...)
}

func (l LogrusLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Entry.WithContext(ctx).Warnf(msg, data...)
}

func (l LogrusLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Entry.WithContext(ctx).Errorf(msg, data...)
}

func (l LogrusLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	switch {
	case err != nil && l.Logger.IsLevelEnabled(logrus.ErrorLevel):
		sql, _ := fc()
		l.Entry.WithContext(ctx).Errorf("%s [%s]", err, sql)
	case elapsed > 200*time.Millisecond && l.Logger.IsLevelEnabled(logrus.WarnLevel):
		sql, _ := fc()
		l.Entry.WithContext(ctx).Warnf("%s [%s]", elapsed, sql)
	case l.Logger.IsLevelEnabled(logrus.DebugLevel):
		sql, _ := fc()
		l.Entry.WithContext(ctx).Debugf("%s [%s]", elapsed, sql)
	}
}
