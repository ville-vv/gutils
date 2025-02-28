package vtask

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestDynamicWorkPool_Submit(t *testing.T) {
	minWorkers := int64(2)
	maxWorkers := int64(20000)
	workPool := NewDynamicWorkPool(WithMinWorkers(minWorkers), WithMaxWorkers(maxWorkers))
	if workPool == nil {
		t.Errorf("NewDynamicWorkPool(%v, %v) = nil", minWorkers, maxWorkers)
	}

	var wg sync.WaitGroup
	isClose := false
	var submitFn = func() {
		defer wg.Done()
		for {
			if isClose {
				return
			}
			for i := 0; i < 30000; i++ {
				err := workPool.Submit(func() {
					//active, processed, queued, errored := workPool.Metrics()
					//fmt.Printf("active: %d, processed: %d, queued: %d errored:%d\n", active, processed, queued, errored)
					//time.Sleep(time.Millisecond * 10)
				})
				if err != nil {
					return
				}
			}
			time.Sleep(time.Second * 1)
		}
	}

	for i := 0; i < 20000; i++ {
		wg.Add(1)
		go submitFn()
	}

	time.Sleep(time.Second * 10)
	isClose = true
	workPool.Release()
	mt := workPool.Metrics()
	fmt.Println("finish active: ", mt.String())
	wg.Wait()
}

func BenchmarkDynamicWorkPool_Submit(b *testing.B) {
	minWorkers := int64(1000)
	maxWorkers := int64(40000)
	workPool := NewDynamicWorkPool(WithMinWorkers(minWorkers), WithMaxWorkers(maxWorkers))
	if workPool == nil {
		fmt.Printf("NewDynamicWorkPool(%v, %v) = nil\n", minWorkers, maxWorkers)
		return
	}

	for i := 0; i < b.N; i++ {
		err := workPool.Submit(func() {
			//time.Sleep(time.Millisecond * 1)
		})
		if err != nil {
			fmt.Println("Submit error:", err)
		}
	}
	workPool.Release()
}

// 可配置参数
const (
	MinComputeSteps = 35  // 最小计算步数（控制CPU负载下限）
	MaxComputeSteps = 100 // 最大计算步数（控制CPU负载上限）
)

// CPU密集型任务算法：计算斐波那契数列的第n项（迭代实现）
func cpuIntensiveTask(n int) int {
	if n <= 0 {
		return 0
	}
	a, b := 0, 1
	for i := 1; i < n; i++ {
		a, b = b, a+b
	}
	return b
}

// 生成随机任务
func generateTask() func() {
	steps := MinComputeSteps + rand.Intn(MaxComputeSteps-MinComputeSteps+1)
	return func() {
		cpuIntensiveTask(steps)
		//start := time.Now()
		//result := cpuIntensiveTask(steps)
		//duration := time.Since(start)
		//fmt.Printf("Task completed in %v | Fib(%d)=%d\n", duration, steps, result)
	}
}

func TestDynamicWorkPool_CPU(t *testing.T) {
	minWorkers := int64(2)
	maxWorkers := int64(20000)
	workPool := NewDynamicWorkPool(WithMinWorkers(minWorkers), WithMaxWorkers(maxWorkers))
	if workPool == nil {
		t.Errorf("NewDynamicWorkPool(%v, %v) = nil", minWorkers, maxWorkers)
	}

	var wg sync.WaitGroup
	isClose := false
	var submitFn = func() {
		defer wg.Done()
		for {
			if isClose {
				return
			}
			for i := 0; i < 30000; i++ {
				err := workPool.Submit(generateTask())
				if err != nil {
					return
				}
			}
			time.Sleep(time.Second * 1)
			mt := workPool.Metrics()
			fmt.Println("Metrics: ", mt.String())
		}
	}
	for i := 0; i < 30000; i++ {
		wg.Add(1)
		go submitFn()
	}
	time.Sleep(time.Second * 5)
	isClose = true
	workPool.Release()
	wg.Wait()
	mt := workPool.Metrics()
	fmt.Println("Finish Metrics: ", mt.String())
}
