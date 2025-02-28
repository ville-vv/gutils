package locks

import (
	"errors"
	"time"
)

var (
	ErrToManyTimes = errors.New("too many times, please try again later")
)

type Locker interface {
	Lock(key string, timeout time.Duration) (string, error)
	UnLock(key string, val string) error
}

// Interceptor 拦截器
type Interceptor interface {
	Intercept(key string, timeout time.Duration) error
}

type Limiter interface {
	Allow(key string, timeout time.Duration) error
}
