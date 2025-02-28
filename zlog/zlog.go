package zlog

import (
	"context"
	jsoniter "github.com/json-iterator/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type LogConfig struct {
	Name       string
	Mode       string `json:"mode"`
	Level      string `json:"level"`
	Path       string `json:"path"`
	InfoFile   string `json:"infoFile"`
	ErrFile    string `json:"errFile"`
	MaxSize    int    `json:"maxSize"`
	MaxBackups int    `json:"maxBackups"`
	MaxAge     int    `json:"maxAge"`
	Compress   bool   `json:"compress"`
	Lumberjack bool   `json:"lumberjack"`
}

func getLogLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// createLogPath 创建文件路径，支持多个文件路径的创建
func openLogFile(fileName string) *os.File {
	if err := os.MkdirAll(filepath.Dir(fileName), 0755); err != nil {
		panic(err)
	}
	logFile, err := os.OpenFile(fileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	return logFile
}

func lumberjackLogger(fileName string, cfg *LogConfig) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   path.Join(cfg.Path, fileName),
		MaxSize:    cfg.MaxSize,    // 每个日志文件的最大大小（MB）
		MaxBackups: cfg.MaxBackups, // 保留旧日志文件的最大数量
		MaxAge:     cfg.MaxAge,     // 保留旧日志文件的最大天数
		Compress:   cfg.Compress,   // 是否压缩/归档旧日志文件
	}
}

func JsonEncode(w io.Writer) zapcore.ReflectedEncoder {
	enc := jsoniter.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

func NewZLog(cfg LogConfig) *zap.Logger {
	encoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		NewReflectedEncoder: JsonEncode, // 使用自定义的json 解析器，速度比较快，zap 默认使用的是 官方json包
		TimeKey:             "time",
		LevelKey:            "level",
		NameKey:             "svc",
		CallerKey:           "call",
		MessageKey:          "msg",
		StacktraceKey:       "stack",                        // 打印栈内容
		EncodeLevel:         zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:          zapcore.ISO8601TimeEncoder,     // ISO8601 UTC 时间格式
		EncodeDuration:      zapcore.SecondsDurationEncoder, //
		EncodeCaller:        zapcore.ShortCallerEncoder,     // 全路径编码器
		EncodeName:          zapcore.FullNameEncoder,
	})

	level := getLogLevel(cfg.Level)
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	if cfg.Mode == "file" {
		if cfg.Path == "" {
			cfg.Path = "./logs"
		}

		if cfg.InfoFile == "" {
			cfg.InfoFile = "access.log"
		}

		if cfg.ErrFile != "" {
			core = newCoreWithErrFile(&cfg, level, encoder)
		} else {
			core = newCoreOnlyOneFile(&cfg, level, encoder)
		}
	}

	return zap.New(core, zap.AddCaller()).Named(cfg.Name)
}

// 自定义本地时区时间编码器
func localTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	// 将时间转换为本地时间，并格式化为 YYYY-MM-DD HH:MM:SS
	enc.AppendString(t.Local().Format("[2006-01-02T15:04:05.000]"))
}

func NewConsoleZLog(cfg LogConfig) *zap.Logger {
	encoder := zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "svc",
		CallerKey:      "call",
		MessageKey:     "msg",
		StacktraceKey:  "stack",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,  // 小写编码器
		EncodeTime:     localTimeEncoder,               // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder, // 秒的持续时间
		EncodeCaller:   zapcore.ShortCallerEncoder,     // 短路径编码器
		//EncodeName: zapcore.FullNameEncoder, // 全路径编码器
		ConsoleSeparator: " ",
	})

	level := getLogLevel(cfg.Level)
	core := zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level)
	if cfg.Mode == "file" {
		if cfg.Path == "" {
			cfg.Path = "./logs"
		}

		if cfg.InfoFile == "" {
			cfg.InfoFile = "z_info.log"
		}

		if cfg.ErrFile != "" {
			core = newCoreWithErrFile(&cfg, level, encoder)
		} else {
			core = newCoreOnlyOneFile(&cfg, level, encoder)
		}
	}

	return zap.New(core, zap.AddCaller()).Named(cfg.Name)
}

func newCoreOnlyOneFile(cfg *LogConfig, level zapcore.Level, encoder zapcore.Encoder) zapcore.Core {
	infoCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(lumberjackLogger(cfg.InfoFile, cfg)),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= level
		}))
	return infoCore
}

func newCoreWithErrFile(cfg *LogConfig, level zapcore.Level, encoder zapcore.Encoder) zapcore.Core {
	infoCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(lumberjackLogger(cfg.InfoFile, cfg)),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel && lvl >= level
		}))
	errCore := zapcore.NewCore(
		encoder,
		zapcore.AddSync(lumberjackLogger(cfg.ErrFile, cfg)),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool { return lvl >= zapcore.ErrorLevel && lvl >= level }))
	return zapcore.NewTee(infoCore, errCore)
}

func getWrite(fileName string, cfg *LogConfig) io.Writer {
	if fileName == "" {
		return os.Stdout
	}
	if cfg.Lumberjack {
		return lumberjackLogger(fileName, cfg)
	}
	return openLogFile(filepath.Join(cfg.Path, fileName))
}

var _log *zap.Logger
var _sugar *zap.SugaredLogger
var _sugarNoSkip *zap.SugaredLogger
var once sync.Once

func init() {
	_log = NewZLog(LogConfig{})
	_sugar = _log.Sugar().WithOptions(zap.AddCallerSkip(1))
	_sugarNoSkip = _log.Sugar()
}

func Init(cfg LogConfig) *zap.Logger {
	once.Do(func() {
		_log = NewZLog(cfg)
		_sugar = _log.Sugar().WithOptions(zap.AddCallerSkip(1))
		_sugarNoSkip = _log.Sugar()
	})
	return _log
}

func InitConsole(cfg ...LogConfig) *zap.Logger {
	once.Do(func() {
		c := LogConfig{}
		if len(cfg) > 0 {
			c = cfg[0]
		}
		_log = NewConsoleZLog(c)
		_sugar = _log.Sugar().WithOptions(zap.AddCallerSkip(1))
		_sugarNoSkip = _log.Sugar()
	})
	return _log
}

func GetLoggerCallSkip(skip int) Logger {
	return _sugar.WithOptions(zap.AddCallerSkip(skip))
}

func GetLoggerWith(kvFields ...interface{}) Logger {
	return _sugarNoSkip.With(kvFields...)
}

func GetLogger() Logger {
	return _sugarNoSkip
}

func Info(args ...interface{}) {
	_sugar.Info(args...)
}

func Infof(template string, args ...interface{}) {
	_sugar.Infof(template, args...)
}

func Infoln(args ...interface{}) {
	_sugar.Infoln(args...)
}

func Infow(msg string, kvs ...interface{}) {
	_sugar.Infow(msg, kvs...)
}

func Debug(args ...interface{}) {
	_sugar.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	_sugar.Debugf(template, args...)
}

func Debugln(args ...interface{}) {
	_sugar.Debugln(args...)
}

func Debugw(msg string, kvs ...interface{}) {
	_sugar.Debugw(msg, kvs...)
}

func Warn(args ...interface{}) {
	_sugar.Warn(args...)
}
func Warnf(template string, args ...interface{}) {
	_sugar.Warnf(template, args...)
}
func Warnln(args ...interface{}) {
	_sugar.Warnln(args...)
}
func Warnw(msg string, kvs ...interface{}) {
	_sugar.Warnw(msg, kvs...)
}
func Error(args ...interface{}) {
	_sugar.Error(args...)
}
func Errorf(template string, args ...interface{}) {
	_sugar.Errorf(template, args...)
}
func Errorln(args ...interface{}) {
	_sugar.Errorln(args...)
}
func Errorw(msg string, kvs ...interface{}) {
	_sugar.Errorw(msg, kvs...)
}

// InfoWithContext 带有 context 的日志方法
func InfoWithContext(ctx context.Context, args ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	_sugar.With("trace", traceID, "span", spanID).Info(args...)
}

func InfofWithContext(ctx context.Context, template string, args ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	_sugar.WithLazy("trace", traceID, "span", spanID).Infof(template, args...)
}

func InfowWithContext(ctx context.Context, msg string, kvs ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	kvs = append(kvs, zap.String("trace", traceID), zap.String("span", spanID))
	_sugar.Infow(msg, kvs...)
}

func DebugWithContext(ctx context.Context, args ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	_sugar.With("trace", traceID, "span", spanID).Debug(args...)
}

func DebugfWithContext(ctx context.Context, template string, args ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	_sugar.With("trace", traceID, "span", spanID).Debugf(template, args...)
}

func DebugwWithContext(ctx context.Context, msg string, kvs ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	kvs = append(kvs, zap.String("trace", traceID), zap.String("span", spanID))
	_sugar.Debugw(msg, kvs...)
}

func WarnWithContext(ctx context.Context, args ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	_sugar.With("trace", traceID, "span", spanID).Warn(args...)
}

func WarnfWithContext(ctx context.Context, template string, args ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	_sugar.With("trace", traceID, "span", spanID).Warnf(template, args...)
}

func WarnwWithContext(ctx context.Context, msg string, kvs ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	kvs = append(kvs, zap.String("trace", traceID), zap.String("span", spanID))
	_sugar.Warnw(msg, kvs...)
}

func ErrorWithContext(ctx context.Context, args ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	_sugar.With("trace", traceID, "span", spanID).Error(args...)
}

func ErrorfWithContext(ctx context.Context, template string, args ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	_sugar.With("trace", traceID, "span", spanID).Errorf(template, args...)
}

func ErrorwWithContext(ctx context.Context, msg string, kvs ...interface{}) {
	traceID, spanID := ExtractTraceInfo(ctx)
	kvs = append(kvs, zap.String("trace", traceID), zap.String("span", spanID))
	_sugar.Errorw(msg, kvs...)
}
