package zlog

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
