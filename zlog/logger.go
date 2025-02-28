package zlog

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// LoggerInterface 定义了所有的日志方法
type Logger interface {
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Infow(msg string, kvs ...interface{})
	Infoln(args ...interface{})
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Debugw(msg string, kvs ...interface{})
	Debugln(args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Warnw(msg string, kvs ...interface{})
	Warnln(args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Errorw(msg string, kvs ...interface{})
	Errorln(args ...interface{})
}

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

func WithContextAndCallerSkip(ctx context.Context) Logger {
	if ctx == nil {
		return &loggerWithContext{ctx: ctx, sugar: _sugar}
	}
	wc := &loggerWithContext{ctx: ctx, sugar: _sugar.WithOptions(zap.AddCallerSkip(1))}
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

func (l *loggerWithContext) extractTraceInfo() (traceID, spanID string) {
	return ExtractTraceInfo(l.ctx)
}

func (l *loggerWithContext) Info(args ...interface{}) {
	if l.ctx == nil {
		l.sugar.Info(args...)
		return
	}
	traceID, spanID := l.extractTraceInfo()
	l.sugar.With("trace", traceID, "span", spanID).Info(args...)
}

func (l *loggerWithContext) Infof(template string, args ...interface{}) {
	if l.ctx == nil {
		l.sugar.Infof(template, args...)
		return
	}
	traceID, spanID := l.extractTraceInfo()
	l.sugar.With("trace", traceID, "span", spanID).Infof(template, args...)
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
	if l.ctx == nil {
		l.sugar.Infoln(args...)
	}
	traceID, spanID := l.extractTraceInfo()
	l.sugar.With("trace", traceID, "span", spanID).Infoln(args...)
}

func (l *loggerWithContext) Debug(args ...interface{}) {
	if l.ctx == nil {
		l.sugar.Debug(args...)
		return
	}
	traceID, spanID := l.extractTraceInfo()
	l.sugar.With("trace", traceID, "span", spanID).Debug(args...)
}

func (l *loggerWithContext) Debugf(template string, args ...interface{}) {
	if l.ctx == nil {
		l.sugar.Debugf(template, args...)
		return
	}
	traceID, spanID := l.extractTraceInfo()
	l.sugar.With("trace", traceID, "span", spanID).Debugf(template, args...)
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
	if l.ctx == nil {
		l.sugar.Debugln(args...)
	}
	traceID, spanID := l.extractTraceInfo()
	l.sugar.With("trace", traceID, "span", spanID).Debugln(args...)
}

func (l *loggerWithContext) Warn(args ...interface{}) {
	if l.ctx == nil {
		l.sugar.Warn(args...)
		return
	}
	l.sugar.With("trace", l.traceID, "span", l.spanID).Warn(args...)
}

func (l *loggerWithContext) Warnf(template string, args ...interface{}) {
	if l.ctx == nil {
		l.sugar.Warnf(template, args...)
		return
	}
	l.sugar.With("trace", l.traceID, "span", l.spanID).Warnf(template, args...)
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
	if l.ctx == nil {
		l.sugar.Warnln(args...)
	}
	traceID, spanID := l.extractTraceInfo()
	l.sugar.With("trace", traceID, "span", spanID).Warnln(args...)
}

func (l *loggerWithContext) Error(args ...interface{}) {
	if l.ctx == nil {
		l.sugar.Error(args...)
		return
	}
	l.sugar.With("trace", l.traceID, "span", l.spanID).Error(args...)
}

func (l *loggerWithContext) Errorf(template string, args ...interface{}) {
	if l.ctx == nil {
		l.sugar.Errorf(template, args...)
		return
	}
	l.sugar.With("trace", l.traceID, "span", l.spanID).Errorf(template, args...)
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
	if l.ctx == nil {
		l.sugar.Errorln(args...)
	}
	traceID, spanID := l.extractTraceInfo()
	l.sugar.With("trace", traceID, "span", spanID).Errorln(args...)
}
