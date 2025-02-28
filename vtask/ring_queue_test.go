package vtask

import (
	"fmt"
	"testing"
)

type asb struct {
	ad int
}

func TestRingQueue_Pop(t *testing.T) {
	rQ := NewRingQueue(40000)
	l := AtomicInt64{}
	add := AtomicInt64{}
	finish := false
	for i := 0; i < 20; i++ {
		go func(n int) {
			for {
				if rQ.Length() == 0 {
					continue
				}
				val := rQ.Pop()
				if val != nil {

					l.Inc()
				}
				if l.Load() >= 20000 {
					finish = true
					fmt.Println("a:", val, l.Load())
					return
				}
			}
		}(i)
	}

	for i := 1; i <= 20; i++ {
		go func(n int) {
			idx := 1
			for {
				//val := vutil.RandStringBytesMask(16)
				err := rQ.Push(&asb{ad: idx * n})
				if err != nil {
					fmt.Println("加入队列出错")
				}
				add.Inc()
				idx++
				if add.Load() >= 20000 {
					return
				}

			}
		}(i)
	}
	for !finish {
	}
	fmt.Println(l.Load(), add.Load())

}

func BenchmarkRingQueue_Push(b *testing.B) {
	rQ := NewRingQueue(int64(b.N))

	for i := 0; i < b.N; i++ {
		rQ.Push(i)
	}
}

func BenchmarkRingQueue_Pop(b *testing.B) {
	b.StopTimer()
	rQ := NewRingQueue(int64(b.N))

	for i := 0; i < b.N; i++ {
		rQ.Push(i)
	}

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		rQ.Pop()
	}
}
