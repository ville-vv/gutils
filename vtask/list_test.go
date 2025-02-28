package vtask

import (
	"fmt"
	"testing"
	"time"
)

func TestQuery_Push2(t *testing.T) {
	po := New()
	num := 10000000
	start := time.Now().UnixNano()
	for i := 0; i < num; i++ {
		po.Push(i)
	}
	end := time.Now().UnixNano()
	fmt.Println("Push的时间：", (end-start)/1e6)

	fmt.Println("长度：", Length())

	start = time.Now().UnixNano()
	for i := 0; i < num; i++ {
		Pop()
	}
	end = time.Now().UnixNano()
	fmt.Println("Pop的时间：", (end-start)/1e6)
}

func BenchmarkStack_Push(b *testing.B) {
	po := New()
	for i := 0; i < b.N; i++ {
		po.Push(i)
	}
}

func BenchmarkStack_Pop(b *testing.B) {
	po := New()
	b.StopTimer()

	for i := 0; i < b.N; i++ {
		po.Push(i)
	}
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		Pop()
	}
}

func TestPush(t *testing.T) {
	Push(30)
	Push(50)
	Push(90)
	Push(100)
	Push(5)

	fmt.Println(Pop())
	fmt.Println(Pop())
	fmt.Println(Pop())
	fmt.Println(Pop())
	fmt.Println(Pop())

	Push(30)
	Push(50)
	Push(90)
	Push(100)
	Push(5)

	fmt.Println(Shift())
	fmt.Println(Shift())
	fmt.Println(Shift())
	fmt.Println(Shift())
	fmt.Println(Shift())
}
