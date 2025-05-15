package zlog

import (
	"go.uber.org/zap"
)

// loggerWithContextTrace 包含 context.Context 和日志输出方法
type loggerWithContextTrace struct {
	traceID string
	sugar   *zap.SugaredLogger
}

func WithTraceID(traceID string) Logger {
	wc := &loggerWithContextTrace{traceID: traceID, sugar: _sugar}
	return wc
}

func (l *loggerWithContextTrace) logger() *zap.SugaredLogger {
	return l.sugar.With("trace", l.traceID)
}

func (l *loggerWithContextTrace) Info(args ...interface{}) {
	l.logger().Info(args...)
}

func (l *loggerWithContextTrace) Infof(template string, args ...interface{}) {
	l.logger().Infof(template, args...)
}

func (l *loggerWithContextTrace) Infow(msg string, keysAndValues ...interface{}) {
	if l.traceID == "" {
		l.sugar.Infow(msg, keysAndValues...)
		return
	}
	keysAndValues = append(keysAndValues, zap.String("trace", l.traceID))
	l.sugar.Infow(msg, keysAndValues...)
}

func (l *loggerWithContextTrace) Infoln(args ...interface{}) {
	l.logger().Infoln(args...)
}

func (l *loggerWithContextTrace) Debug(args ...interface{}) {
	l.logger().Debug(args...)
}

func (l *loggerWithContextTrace) Debugf(template string, args ...interface{}) {
	l.logger().Debugf(template, args...)
}

func (l *loggerWithContextTrace) Debugw(msg string, keysAndValues ...interface{}) {
	if l.traceID == "" {
		l.sugar.Debugw(msg, keysAndValues...)
		return
	}
	l.sugar.Debugw(msg, append(keysAndValues, zap.String("trace", l.traceID))...)
}

func (l *loggerWithContextTrace) Debugln(args ...interface{}) {
	l.logger().Debugln(args...)
}

func (l *loggerWithContextTrace) Warn(args ...interface{}) {
	l.logger().Warn(args...)
}

func (l *loggerWithContextTrace) Warnf(template string, args ...interface{}) {
	l.logger().Warnf(template, args...)
}

func (l *loggerWithContextTrace) Warnw(msg string, keysAndValues ...interface{}) {
	if l.traceID == "" {
		l.sugar.Warnw(msg, keysAndValues...)
		return
	}
	l.sugar.Warnw(msg, append(keysAndValues, zap.String("trace", l.traceID))...)
}

func (l *loggerWithContextTrace) Warnln(args ...interface{}) {
	l.logger().Warnln(args...)
}

func (l *loggerWithContextTrace) Error(args ...interface{}) {
	l.logger().Error(args...)
}

func (l *loggerWithContextTrace) Errorf(template string, args ...interface{}) {
	l.logger().Errorf(template, args...)
}

func (l *loggerWithContextTrace) Errorw(msg string, keysAndValues ...interface{}) {
	if l.traceID == "" {
		l.sugar.Errorw(msg, keysAndValues...)
		return
	}
	l.sugar.Errorw(msg, append(keysAndValues, zap.String("trace", l.traceID))...)
}

func (l *loggerWithContextTrace) Errorln(args ...interface{}) {
	l.logger().Errorln(args...)
}
