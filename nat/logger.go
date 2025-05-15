package nat

var _log Logger = &defaultLogger{}

type Logger interface {
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Warnf(format string, args ...interface{})
}

func SetLogger(l Logger) {
	if l == nil {
		l = &defaultLogger{}
	}
	_log = l
}

type defaultLogger struct{}

func (l *defaultLogger) Infof(format string, args ...interface{}) {
	// default not anything implemented
}

func (l *defaultLogger) Errorf(format string, args ...interface{}) {
	// default not anything implemented
}

func (l *defaultLogger) Warnf(format string, args ...interface{}) {
	// default not anything implemented
}
