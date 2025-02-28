package vtask

import "sync/atomic"

type AtomicInt64 struct {
	num int64
}

func (sel *AtomicInt64) Inc() {
	atomic.AddInt64(&sel.num, 1)
}

func (sel *AtomicInt64) Dec() {
	atomic.AddInt64(&sel.num, -1)
}

func (sel *AtomicInt64) Load() int64 {
	return atomic.LoadInt64(&sel.num)
}

func (sel *AtomicInt64) Store(n int64) {
	atomic.StoreInt64(&sel.num, n)
}
