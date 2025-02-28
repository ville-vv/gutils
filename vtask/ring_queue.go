package vtask

import (
	"sync"
)

type RingQueue struct {
	lock   sync.Mutex
	list   []interface{}
	qLen   int64
	qCap   int64
	ptrIdx int64 // 指定当前元素存放位置
	popIdx int64
}

func NewRingQueue(qCap int64) *RingQueue {
	if qCap <= 0 {
		qCap = 60000
	}
	return &RingQueue{
		list:   make([]interface{}, qCap+1),
		qCap:   qCap,
		ptrIdx: 0,
		popIdx: 0,
	}
}

func (r *RingQueue) Push(val interface{}) error {
	r.lock.Lock()
	if r.qLen > r.qCap {
		r.lock.Unlock()
		return ErrOverMaxSize
	}
	r.list[r.ptrIdx] = val
	r.qLen++
	r.ptrIdx++
	if r.ptrIdx >= r.qCap {
		r.ptrIdx = 0
	}
	r.lock.Unlock()
	return nil
}

func (r *RingQueue) Pop() interface{} {
	r.lock.Lock()
	if r.qLen == 0 {
		r.lock.Unlock()
		return nil
	}
	val := r.list[r.popIdx]
	r.list[r.popIdx] = nil
	r.qLen--
	r.popIdx++
	if r.popIdx >= r.qCap {
		r.popIdx = 0
	}
	r.lock.Unlock()
	return val
}

func (r *RingQueue) Length() int64 {
	r.lock.Lock()
	l := r.qLen
	r.lock.Unlock()
	return l
}
