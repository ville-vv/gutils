package vtask

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type TaskID uint64

type TaskHandler func(interface{})

const (
	slotBits = 24 // 槽位索引占用位数
	seqBits  = 40 // 序列号位数
	maxSlots = 1<<slotBits - 1
	slotMask = 0xFFFFFF0000000000
	seqMask  = 0x000000FFFFFFFFFF
)

type taskEntry struct {
	id        TaskID      // 编码后的任务ID
	param     interface{} // 任务参数
	handler   TaskHandler // 任务处理函数
	executeAt int64       // 执行时间戳（纳秒）
}

// Slot 新增结构体封装槽相关数据
type Slot struct {
	id    int          // 槽位ID
	tasks *list.List   // 槽内任务链表
	lock  sync.RWMutex // 槽级锁
}

func (s *Slot) PushBack(entry *taskEntry) *list.Element {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.tasks.PushBack(entry)
}

func (s *Slot) Remove(elem *list.Element) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.tasks.Remove(elem)
}

// prepareTasks 优化点1: 快速获取待处理任务（最小化锁时间）
func (s *Slot) prepareTasks() []*taskEntry {
	now := time.Now().UnixNano()
	var tasks []*taskEntry
	s.lock.Lock()
	for elem := s.tasks.Front(); elem != nil; {
		entry := elem.Value.(*taskEntry)
		next := elem.Next()
		if entry.executeAt <= now {
			tasks = append(tasks, entry)
			s.tasks.Remove(elem)
		}
		elem = next
	}
	s.lock.Unlock()
	return tasks
}

type TimeWheelOption func(tw *timeWheelConfig)

type timeWheelConfig struct {
	interval time.Duration
	slotsNum int
}

func (sel *timeWheelConfig) getInterval() time.Duration {
	if sel.interval == 0 {
		return time.Second
	}
	return sel.interval
}

func (sel *timeWheelConfig) getSlotsNum() int {
	if sel.slotsNum == 0 {
		return 60
	}
	return sel.slotsNum
}

func WithTimeWheelInterval(interval time.Duration) TimeWheelOption {
	return func(tw *timeWheelConfig) {
		tw.interval = interval
	}
}

func WithTimeWheelSlotsNum(slotsNum int) TimeWheelOption {
	if slotsNum > maxSlots {
		slotsNum = maxSlots
	}
	return func(tw *timeWheelConfig) {
		tw.slotsNum = slotsNum
	}
}

type TimeWheel struct {
	interval   time.Duration // 时间间隔
	taskLen    int64         // 任务数量
	slotNum    int           // 槽位数量
	slots      []*Slot       // 时间槽链表
	cursor     int           // 当前槽指针
	ticker     *time.Ticker  // 时间驱动器
	taskMap    sync.Map      // 任务存储 map[TaskID]*list.Element
	idSequence atomic.Uint64 // 原子ID生成器
	stopCh     chan struct{}
	wg         sync.WaitGroup
	workPool   *DynamicWorkPool
}

func NewTimeWheel(opts ...TimeWheelOption) *TimeWheel {
	cfg := &timeWheelConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	interval := cfg.getInterval()
	slotsNum := cfg.getSlotsNum()

	slots := make([]*Slot, slotsNum)
	for i := range slots {
		slots[i] = &Slot{
			id:    i,
			tasks: list.New(),
		}
	}

	tw := &TimeWheel{
		interval: interval,
		slotNum:  slotsNum,
		slots:    slots,
		cursor:   0,
		stopCh:   make(chan struct{}),
	}

	return tw
}

func (tw *TimeWheel) Start() {
	tw.ticker = time.NewTicker(tw.interval)
	tw.wg.Add(1)
	tw.workPool = NewDynamicWorkPool(WithMinWorkers(1), WithMaxWorkers(1000))
	go func() {
		defer tw.wg.Done()
		for {
			select {
			case <-tw.ticker.C:
				tw.advance()
			case <-tw.stopCh:
				tw.ticker.Stop()
				return
			}
		}
	}()
}

// 生成带槽位信息的任务ID
func (tw *TimeWheel) generateID(slotIdx int) TaskID {
	seq := tw.idSequence.Add(1)
	return TaskID((uint64(slotIdx) << seqBits) | (seq & seqMask))
}

// 解析槽位索引
func (tw *TimeWheel) slotFromID(id TaskID) int {
	return int((uint64(id) & slotMask) >> seqBits)
}

func (tw *TimeWheel) AddTask(delay time.Duration, handler TaskHandler, param interface{}) TaskID {
	// 计算目标槽位
	steps := int(delay / tw.interval)
	slotIdx := (tw.cursor + steps) % tw.slotNum
	// 生成唯一ID
	id := tw.generateID(slotIdx)
	entry := &taskEntry{
		id:        id,
		param:     param,
		handler:   handler,
		executeAt: time.Now().Add(delay).UnixNano(),
	}
	// 槽级锁控制 插入链表并记录元素
	elem := tw.slots[slotIdx].PushBack(entry)
	// 原子操作存储任务
	atomic.AddInt64(&tw.taskLen, 1)
	tw.taskMap.Store(id, elem)
	return id
}

func (tw *TimeWheel) advance() {
	// 1. 快速获取任务快照
	tasks := tw.slots[tw.cursor].prepareTasks()
	tw.cursor = (tw.cursor + 1) % tw.slotNum
	// 2. 并行处理任务
	tw.processTasks(tasks)
}

// 优化后的协程池处理
func (tw *TimeWheel) processTasks(tasks []*taskEntry) {
	if len(tasks) == 0 {
		return
	}
	for _, task := range tasks {
		atomic.AddInt64(&tw.taskLen, -1)
		_ = tw.workPool.Submit(func() {
			tw.safeExecute(task)
		})
	}
}

// 批量处理任务
func (tw *TimeWheel) processBatch(batch []*taskEntry) {
	for _, entry := range batch {
		tw.safeExecute(entry)
	}
}

// 优化点4: 安全执行（含Recover）
func (tw *TimeWheel) safeExecute(entry *taskEntry) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("task panic: %v\n", r)
		}
	}()
	tw.wg.Add(1)
	defer tw.wg.Done()
	entry.handler(entry.param)
}

func (tw *TimeWheel) RemoveTask(id TaskID) {
	slotIdx := tw.slotFromID(id)
	// 原子操作获取并删除（优化点2）
	elem, loaded := tw.taskMap.LoadAndDelete(id)
	if !loaded {
		return
	}
	// 类型断言安全检查（优化点3）
	if listElem, ok := elem.(*list.Element); ok {
		tw.slots[slotIdx].Remove(listElem)
	}
}

func (tw *TimeWheel) Len() int64 {
	return atomic.LoadInt64(&tw.taskLen)
}

func (tw *TimeWheel) Stop() {
	close(tw.stopCh)
	tw.wg.Wait()
	if tw.workPool != nil {
		tw.workPool.Release()
	}
}
