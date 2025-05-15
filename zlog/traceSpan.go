package zlog

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

func ExtractTraceInfo(ctx context.Context) (traceID, spanID string) {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.HasTraceID() {
		traceID = spanCtx.TraceID().String()
	}
	if spanCtx.HasSpanID() {
		spanID = spanCtx.SpanID().String()
	}
	return
}

// loggerWithContext 包含 context.Context 和日志输出方法
type loggerWithContext struct {
	ctx     context.Context
	traceID string
	spanID  string
	sugar   *zap.SugaredLogger
}

func WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return &loggerWithContext{ctx: ctx, sugar: _sugar}
	}
	wc := &loggerWithContext{ctx: ctx, sugar: _sugar}
	wc.traceID, wc.spanID = ExtractTraceInfo(ctx)
	return wc
}

func WithContextPointLogger(ctx context.Context, sugar *zap.SugaredLogger) Logger {
	if ctx == nil {
		return &loggerWithContext{ctx: ctx, sugar: sugar}
	}
	wc := &loggerWithContext{ctx: ctx, sugar: sugar}
	wc.traceID, wc.spanID = ExtractTraceInfo(ctx)
	return wc
}

func (l *loggerWithContext) logger() *zap.SugaredLogger {
	if l.ctx == nil {
		return l.sugar
	}
	traceID, spanID := l.extractTraceInfo()
	return l.sugar.With("trace", traceID, "span", spanID)
}

func (l *loggerWithContext) extractTraceInfo() (traceID, spanID string) {
	if l.traceID != "" || l.spanID != "" {
		return l.traceID, l.spanID
	}
	return ExtractTraceInfo(l.ctx)
}

func (l *loggerWithContext) Info(args ...interface{}) {
	l.logger().Info(args...)
}

func (l *loggerWithContext) Infof(template string, args ...interface{}) {
	l.logger().Infof(template, args...)
}

func (l *loggerWithContext) Infow(msg string, keysAndValues ...interface{}) {
	if l.ctx == nil {
		l.sugar.Infow(msg, keysAndValues...)
		return
	}
	keysAndValues = append(keysAndValues, zap.String("trace", l.traceID), zap.String("span", l.spanID))
	l.sugar.Infow(msg, keysAndValues...)
}

func (l *loggerWithContext) Infoln(args ...interface{}) {
	l.logger().Infoln(args...)
}

func (l *loggerWithContext) Debug(args ...interface{}) {
	l.logger().Debug(args...)
}

func (l *loggerWithContext) Debugf(template string, args ...interface{}) {
	l.logger().Debugf(template, args...)
}

func (l *loggerWithContext) Debugw(msg string, keysAndValues ...interface{}) {
	if l.ctx == nil {
		l.sugar.Debugw(msg, keysAndValues...)
		return
	}
	keysAndValues = append(keysAndValues, zap.String("trace", l.traceID), zap.String("span", l.spanID))
	l.sugar.Debugw(msg, keysAndValues...)
}

func (l *loggerWithContext) Debugln(args ...interface{}) {
	l.logger().Debugln(args...)
}

func (l *loggerWithContext) Warn(args ...interface{}) {
	l.logger().Warn(args...)
}

func (l *loggerWithContext) Warnf(template string, args ...interface{}) {
	l.logger().Warnf(template, args...)
}

func (l *loggerWithContext) Warnw(msg string, keysAndValues ...interface{}) {
	if l.ctx == nil {
		l.sugar.Warnw(msg, keysAndValues...)
		return
	}
	keysAndValues = append(keysAndValues, zap.String("trace", l.traceID), zap.String("span", l.spanID))
	l.sugar.Warnw(msg, keysAndValues...)
}

func (l *loggerWithContext) Warnln(args ...interface{}) {
	l.logger().Warnln(args...)
}

func (l *loggerWithContext) Error(args ...interface{}) {
	l.logger().Error(args...)
}

func (l *loggerWithContext) Errorf(template string, args ...interface{}) {
	l.logger().Errorf(template, args...)
}

func (l *loggerWithContext) Errorw(msg string, keysAndValues ...interface{}) {
	if l.ctx == nil {
		l.sugar.Errorw(msg, keysAndValues...)
		return
	}
	keysAndValues = append(keysAndValues, zap.String("trace", l.traceID), zap.String("span", l.spanID))
	l.sugar.Errorw(msg, keysAndValues...)
}

func (l *loggerWithContext) Errorln(args ...interface{}) {
	l.logger().Errorln(args...)
}
