package zlog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"os"
	"testing"
)

var zlog = NewZLog(LogConfig{
	Name:       "gateway",
	Mode:       "file",
	Level:      "info",
	Path:       "./logggers",
	MaxSize:    1024,
	MaxBackups: 10,
	MaxAge:     30,
	Compress:   true,
})

// InitLogger initializes the zap logger
func InitLogger() *zap.Logger {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	// 设置日志输出文件
	//logFile, err := os.Create("zaplog.log")
	//if err != nil {
	//	panic(err)
	//}

	writeSyncer := zapcore.AddSync(os.Stdout)
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config.EncoderConfig),
		writeSyncer,
		config.Level,
	)

	logger := zap.New(core)
	zap.ReplaceGlobals(logger)
	return logger
}

func InitLoggerA() zerolog.Logger {
	//logFile, err := os.Create("zeroLog.log")
	//if err != nil {
	//	panic(err)
	//}

	logger := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
	return logger
}

func InitLoggerZ() zerolog.Logger {
	logFile := &lumberjack.Logger{
		Filename:   "zeroLog.log",
		MaxSize:    10,   // 日志文件最大大小（单位：MB）
		MaxBackups: 5,    // 保留旧日志文件的最大数量
		MaxAge:     30,   // 保留旧日志文件的最大天数
		Compress:   true, // 是否压缩/归档旧日志文件
	}

	//multi := zerolog.MultiLevelWriter(os.Stdout, logFile)
	logger := zerolog.New(logFile).With().Timestamp().Logger()
	return logger
}

var jsonObj = map[string]interface{}{
	"name": "name23",
	"age":  30,
	"address": map[string]interface{}{
		"city":  "shenzhen",
		"email": "shenzhen@qq.com",
		"eNum":  12344,
	},
}

func TestJsonIndent(t *testing.T) {

	jsonBytes, err := json.Marshal(jsonObj)
	assert.NoError(t, err)

	var out bytes.Buffer
	err = json.Indent(&out, jsonBytes, "", "  ")
	assert.NoError(t, err)
	fmt.Println(string(out.Bytes()))

	zapLog := InitLogger()
	zeroLog := InitLoggerA()
	zapLog.Info("hello world", zap.Any("jsonObj", jsonObj))
	zlog.Info("hello work logx:", zap.Any("jsonObj", jsonObj))
	zeroLog.Info().Any("body", jsonObj).Msg("")
	zapLog.Info("hello world", zap.Any("jsonObj", jsonObj))
	zapLog.Sugar().Infow("hello world")
	//zlog.Error("hello world", "badfa", 1234, "sdfsd", jsonObj)
}

func BenchmarkNewZLog(b *testing.B) {
	b.StopTimer()
	zaplog := InitLogger()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		zaplog.Info("hello world", zap.Any("jsonObj", jsonObj))
	}
}

func BenchmarkNewZLogInfoLn(b *testing.B) {
	b.StopTimer()
	zaplog := InitLogger()
	b.StartTimer()
	l := zaplog.Sugar()
	for i := 0; i < b.N; i++ {
		l.Infoln("hello world", jsonObj)
	}
}

func BenchmarkNewZLogInfo(b *testing.B) {
	b.StopTimer()
	zaplog := InitLogger()
	b.StartTimer()

	l := zaplog.Sugar()
	for i := 0; i < b.N; i++ {
		l.Infow("hello world", "badfa", 1234, "sdfsd", jsonObj)
	}
}
