package vtask

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"
)

var (
	ErrPoolClosed    = errors.New("pool is closed")
	ErrSubmitTimeout = errors.New("submit timeout")
)

type Task func()

type DynamicWorkOption func(*DynamicWorkOptions)

type DynamicWorkOptions struct {
	minWorkers     int64
	maxWorkers     int64
	manageInterval time.Duration
}

func getDynamicWorkOptions(opts ...DynamicWorkOption) *DynamicWorkOptions {
	options := &DynamicWorkOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

func (sel *DynamicWorkOptions) GetMinWorkers() int64 {
	return sel.minWorkers
}

func (sel *DynamicWorkOptions) GetMaxWorkers() int64 {
	if sel.maxWorkers <= 0 {
		sel.maxWorkers = 1000
	}
	return sel.maxWorkers
}

func (sel *DynamicWorkOptions) GetManageInterval() time.Duration {
	if sel.manageInterval <= 0 {
		sel.manageInterval = 1 * time.Second
	}
	return sel.manageInterval
}

func WithMinWorkers(minWorkers int64) DynamicWorkOption {
	return func(sel *DynamicWorkOptions) {
		sel.minWorkers = minWorkers
	}
}

func WithMaxWorkers(maxWorkers int64) DynamicWorkOption {
	return func(sel *DynamicWorkOptions) {
		sel.maxWorkers = maxWorkers
	}
}

func WithManageInterval(val time.Duration) DynamicWorkOption {
	return func(sel *DynamicWorkOptions) {
		sel.manageInterval = val
	}
}

type DynamicWorkPool struct {
	minWorkers     int64
	maxWorkers     int64
	taskQueueCap   int64
	activeWorkers  int64         // 当前活跃的工作协程数量
	submitTasks    int64         // 提交任务数
	processedTasks int64         // 已处理任务数
	queueLength    int64         // 队列长度
	submitErrs     int64         // 提交错误数
	manageInterval time.Duration // 管理间隔时间
	taskQueue      chan Task     // 存放任务的队列
	adjustChan     chan struct{} // 调整信号通道
	workerStopCh   chan struct{} // 用来控制工作协程数量
	stopCh         chan struct{} //
	wg             sync.WaitGroup
	closeOnce      sync.Once
}

func NewDynamicWorkPool(opts ...DynamicWorkOption) *DynamicWorkPool {
	sel := &DynamicWorkPool{
		stopCh:     make(chan struct{}),
		adjustChan: make(chan struct{}, 1),
	}

	options := getDynamicWorkOptions(opts...)
	sel.minWorkers = options.GetMinWorkers()
	sel.maxWorkers = options.GetMaxWorkers()
	sel.manageInterval = options.GetManageInterval()
	if sel.maxWorkers < sel.minWorkers {
		sel.maxWorkers = sel.minWorkers * 2
	}
	sel.taskQueueCap = sel.maxWorkers * 2
	sel.taskQueue = make(chan Task, sel.taskQueueCap)
	sel.workerStopCh = make(chan struct{}, sel.maxWorkers)
	for i := 0; i < int(sel.minWorkers); i++ {
		sel.addWorker()
	}
	sel.wg.Add(1)
	go sel.manager()
	return sel
}

func (sel *DynamicWorkPool) Submit(task Task) error {
	return sel.SubmitWithTimeout(task, 0)
}

func (sel *DynamicWorkPool) entryTask(task Task) bool {
	select {
	case <-sel.stopCh:
	default:
	}
	select {
	case sel.taskQueue <- task:
		atomic.AddInt64(&sel.queueLength, 1)
		atomic.AddInt64(&sel.submitTasks, 1)
		return true
	default:
		return false
	}
}

func (sel *DynamicWorkPool) SubmitWithTimeout(task Task, timeout time.Duration) error {
	if sel.isStop() {
		atomic.AddInt64(&sel.submitErrs, 1)
		return ErrPoolClosed
	}
	ok := sel.entryTask(task)
	if ok {
		return nil
	}
	sel.triggerAdjust()
	err := sel.waitForSubmit(task, timeout)
	if err != nil {
		atomic.AddInt64(&sel.submitErrs, 1)
		return err
	}
	atomic.AddInt64(&sel.queueLength, 1)
	atomic.AddInt64(&sel.submitTasks, 1)
	return nil
}

// 等待任务提交
func (sel *DynamicWorkPool) waitForSubmit(task Task, timeout time.Duration) error {
	if timeout > 0 {
		// 设定超时时间
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case <-sel.stopCh:
			return ErrPoolClosed
		case <-timer.C:
			return ErrSubmitTimeout
		case sel.taskQueue <- task:
		}
		return nil
	}
	// 无超时时间
	select {
	case <-sel.stopCh:
		return ErrPoolClosed
	default:
		select {
		case sel.taskQueue <- task:
		}
	}
	return nil
}

func (sel *DynamicWorkPool) SubmitWithRetry(task func(), maxRetries int, timeout time.Duration) error {
	if timeout <= 0 {
		return fmt.Errorf("timeout must be greater than 0")
	}
	for i := 0; i < maxRetries; i++ {
		err := sel.SubmitWithTimeout(task, timeout)
		if err == nil || errors.Is(err, ErrPoolClosed) {
			return err
		}
		time.Sleep(time.Duration(i*50) * time.Millisecond)
	}
	return fmt.Errorf("submit failed after %d retries", maxRetries)
}

// 触发即时调整
func (sel *DynamicWorkPool) triggerAdjust() bool {
	select {
	case <-sel.stopCh:
		return false
	default:
	}
	select {
	case sel.adjustChan <- struct{}{}:
		return true
	default:
		return false
	}
}

func (sel *DynamicWorkPool) manager() {
	defer sel.wg.Done()
	ticker := time.NewTicker(sel.manageInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			sel.autoAdjust()
		case <-sel.adjustChan: // 即时触发调整
			sel.autoAdjust()
		case <-sel.stopCh:
			return
		}
	}
}

// autoAdjust 自动调整
func (sel *DynamicWorkPool) autoAdjust() {
	current := atomic.LoadInt64(&sel.activeWorkers)
	desired := sel.calculateDesiredWorkers()
	// 扩容逻辑
	if desired > current {
		sel.scaleUp(int(desired - current))
	}
	// 缩容逻辑（保持原有策略）
	if desired < current {
		sel.scaleDown(int(current - desired))
	}
}

func (sel *DynamicWorkPool) scaleUp(num int) {
	maxAdd := int(sel.maxWorkers - atomic.LoadInt64(&sel.activeWorkers))
	add := int(math.Min(float64(num), float64(maxAdd)))
	for i := 0; i < add; i++ {
		sel.addWorker()
	}
}

// 自定义 clamp 函数
func (sel *DynamicWorkPool) clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func (sel *DynamicWorkPool) calculateDesiredWorkers() int64 {
	const (
		smoothFactor = 6.0 // 增大平滑系数，控制曲线陡峭度
	)
	queueLen := atomic.LoadInt64(&sel.queueLength)
	queueCap := float64(sel.taskQueueCap)
	currentQueue := float64(queueLen)

	// 当队列容量为0时直接返回最小值（防御性编程）
	if queueCap == 0 || queueLen == 0 {
		return sel.minWorkers
	}
	// 计算归一化输入 [-smoothFactor, +smoothFactor]
	normalizedInput := (currentQueue/queueCap)*2*smoothFactor - smoothFactor
	// 压力值计算
	pressure := 1 / (1 + math.Exp(-normalizedInput))
	// 线性映射到[minWorkers, maxWorkers]
	desired := float64(sel.minWorkers) + (float64(sel.maxWorkers)-float64(sel.minWorkers))*pressure
	// 边界约束
	return int64(sel.clamp(desired, float64(sel.minWorkers), float64(sel.maxWorkers)))
}

func (sel *DynamicWorkPool) addWorker() {
	sel.wg.Add(1)
	atomic.AddInt64(&sel.activeWorkers, 1)
	go sel.runNewWorker()
}

func (sel *DynamicWorkPool) runNewWorker() {
	defer func() {
		atomic.AddInt64(&sel.activeWorkers, -1)
		sel.wg.Done()
	}()
	for {
		select {
		case task, ok := <-sel.taskQueue:
			if !ok {
				return
			}
			task()
			atomic.AddInt64(&sel.queueLength, -1)
			atomic.AddInt64(&sel.processedTasks, 1)
		case <-sel.workerStopCh:
			return
		case <-sel.stopCh:
			return
		}
	}
}

func (sel *DynamicWorkPool) scaleDown(num int) {
	current := atomic.LoadInt64(&sel.activeWorkers)
	remove := int(math.Min(float64(num), float64(current-sel.minWorkers)))
	for i := 0; i < remove; i++ {
		select {
		case sel.workerStopCh <- struct{}{}:
		default:
			return
		}
	}
}

func (sel *DynamicWorkPool) Release() {
	sel.closeOnce.Do(func() {
		close(sel.stopCh)
		sel.wg.Wait() // 等待工作协程完全退出
		close(sel.taskQueue)
		close(sel.workerStopCh)
		close(sel.adjustChan)
	})
}

func (sel *DynamicWorkPool) isStop() bool {
	select {
	case <-sel.stopCh:
		return true
	default:
		return false
	}
}

func (sel *DynamicWorkPool) ReleaseWait() {
	sel.closeOnce.Do(func() {
		close(sel.stopCh)
		sel.waitFinished()
		close(sel.taskQueue)
		close(sel.workerStopCh)
		close(sel.adjustChan)
		sel.wg.Wait() // 等待工作协程完全退出
	})
}

func (sel *DynamicWorkPool) waitFinished() {
	for atomic.LoadInt64(&sel.queueLength) != 0 {
		time.Sleep(time.Millisecond * 100)
	}
}

type DynamicWorkPoolMetrics struct {
	TotalTasks    int
	ActiveWorkers int
	Processed     int
	Queued        int
}

func (sel *DynamicWorkPoolMetrics) String() string {
	return fmt.Sprintf("total: %d, processed: %d, active: %d, queued: %d",
		sel.TotalTasks, sel.Processed, sel.ActiveWorkers, sel.Queued)
}

func (sel *DynamicWorkPool) Metrics() DynamicWorkPoolMetrics {
	mt := DynamicWorkPoolMetrics{
		TotalTasks:    int(atomic.LoadInt64(&sel.submitTasks)),
		ActiveWorkers: int(atomic.LoadInt64(&sel.activeWorkers)),
		Processed:     int(atomic.LoadInt64(&sel.processedTasks)),
		Queued:        int(atomic.LoadInt64(&sel.queueLength)),
	}
	return mt
}
