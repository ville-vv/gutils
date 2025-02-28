// @File     : task
// @Author   : Ville
// @Time     : 19-10-9 上午9:08
// vtask
package vtask

//
//import (
//	"context"
//	"errors"
//	"sync"
//	"sync/atomic"
//	"time"
//)
//
//type TaskFunc func(params interface{}) error
//
//type element struct {
//	params interface{}
//	times  int
//	tf     TaskFunc
//}
//
//func (e *element) IncTimes() {
//	e.times += 1
//}
//
//type Task struct {
//	ps      bool          // 任务是否启动了
//	scheme  string        // 方案名称
//	runCnt  int32         // 当前任务数量
//	ts      chan element  // 任务执行的单元
//	inl     time.Duration // 对执行错误的任务循环处理一次的定时
//	reLoops int           //
//	cache   Queue         //
//}
//
//func NewTask(num ...int) *Task {
//	n := 10
//	if len(num) > 0 {
//		n = num[0]
//	}
//	return &Task{
//		ts:      make(chan element, n),
//		cache:   New(),
//		inl:     time.Second * 5,
//		reLoops: 3,
//	}
//}
//
//// Process is task process
//func (t *Task) Process(ctx context.Context) {
//	if t.ps {
//		return
//	}
//
//	go t.loopCache(ctx)
//	go t.process(ctx)
//}
//
//func (t *Task) process(ctx context.Context) {
//	t.ps = true
//	for {
//		select {
//		case elm, ok := <-t.ts:
//			if !ok {
//				t.ps = false
//				return
//			}
//			if err := elm.tf(elm.params); err != nil {
//				//任务执行失败
//				if err = t.cache.Push(elm); err == nil {
//					break
//				}
//			}
//			atomic.AddInt32(&t.runCnt, -1)
//
//		case <-ctx.Done():
//			t.ps = false
//			return
//		}
//	}
//}
//
//func (t *Task) digestCache() {
//	reloops := make([]element, 0, t.cache.Length())
//	for t.cache.Length() > 0 {
//		nd := t.cache.Pop()
//		if nd != nil {
//			switch elem := nd.(type) {
//			case element:
//				elem.IncTimes()
//				if err := elem.tf(elem.params); err != nil {
//					if elem.times < t.reLoops {
//						// 执行错误再次重试
//						reloops = append(reloops, elem)
//						break
//					}
//				}
//				atomic.AddInt32(&t.runCnt, -1)
//			}
//		}
//	}
//
//	// 重新放入缓存列表中
//	for i := range reloops {
//		_ = t.cache.Push(reloops[i])
//	}
//}
//
//func (t *Task) loopCache(ctx context.Context) {
//	tkr := time.NewTicker(t.inl)
//	for {
//		t.digestCache()
//		select {
//		case <-ctx.Done():
//			return
//		case <-tkr.C:
//		}
//	}
//}
//
//// Add a task , the param will pass on to tf
//func (t *Task) Add(param interface{}, tf TaskFunc) {
//	atomic.AddInt32(&t.runCnt, 1)
//	t.ts <- element{params: param, tf: tf}
//}
//
//// CanStop return true is can stop task
//func (t *Task) CanStop() bool {
//	return atomic.LoadInt32(&t.runCnt) <= 0
//}
//
//// 重试机制
//// 执行循序
//// 分布式全局锁
//// 可持久存储
//
//// 持久化存储
//type Persistent interface {
//	Load() ([]interface{}, error)
//	Store(interface{}) error
//}
//
//type TaskOption struct {
//	TaskExeNum      int   // 任务执行现成数
//	MaxQueueNum     int64 // 最列表务数
//	RetryFlag       bool  // 重试开关
//	Persistent      Persistent
//	NewRetry        func(val interface{}) RetryElem
//	ErrEventHandler func(ctx interface{}, err error)
//	Exec            func(val interface{}) (retry bool)
//}
//
////
//type RetryElem interface {
//	Can() bool
//	Interval() int64
//	GetData() interface{}
//}
//
//type MiniTask struct {
//	taskExeNum      int // 任务执行现成数
//	dataList        Queue
//	retryList       *DelayQueue // 延迟重试
//	retryCh         chan interface{}
//	pst             Persistent // 是否持久化
//	once            sync.Once
//	isStop          bool
//	stopCh          chan bool
//	RetryFlag       bool // 重试开关
//	newRetry        func(val interface{}) RetryElem
//	exec            func(val interface{}) (retry bool)
//	ErrEventHandler func(ctx interface{}, err error)
//}
//
//func NewMiniTask(option *TaskOption) *MiniTask {
//	mtsk := &MiniTask{
//		dataList:        NewRingQueue(option.MaxQueueNum),
//		retryCh:         make(chan interface{}, 2000),
//		pst:             option.Persistent,
//		once:            sync.Once{},
//		RetryFlag:       option.RetryFlag,
//		newRetry:        option.NewRetry,
//		exec:            option.Exec,
//		ErrEventHandler: option.ErrEventHandler,
//		stopCh:          make(chan bool),
//	}
//	mtsk.taskExeNum = option.TaskExeNum
//	if mtsk.taskExeNum == 0 {
//		mtsk.taskExeNum = 5
//	}
//	mtsk.retryList = NewDelayQueue(mtsk.delayPush)
//	return mtsk
//}
//
//func (t *MiniTask) Start() {
//	t.once.Do(func() {
//		var wait sync.WaitGroup
//		t.retryList.Run()
//		wait.Add(t.taskExeNum)
//		for i := 0; i < t.taskExeNum; i++ {
//			go func() {
//				wait.Done()
//				t.loopExec(t.stopCh)
//			}()
//		}
//		wait.Wait()
//		//go func() {
//		//	wait.Done()
//		//	t.loopExec(t.stopCh)
//		//}()
//		wait.Add(1)
//		go func() {
//			wait.Done()
//			t.loopRetry()
//		}()
//		wait.Wait()
//		err := t.persistentLoad()
//		if err != nil {
//			panic(err)
//		}
//	})
//}
//
//func (t *MiniTask) Stop() {
//	if t.isStop {
//		return
//	}
//	t.isStop = true
//	t.retryList.Close()
//	t.waitStop()
//}
//
//func (t *MiniTask) waitStop() {
//	t.clearRetryCh()
//	t.clearQueue()
//}
//
//var (
//	RetryTimeout = errors.New("retry is out times")
//)
//
//func (t *MiniTask) retryPush(retryDt RetryElem) {
//	if !retryDt.Can() {
//		t.ErrEventHandler(retryDt.GetData(), RetryTimeout)
//		return
//	}
//	if err := t.retryList.Push(retryDt, retryDt.Interval()); err != nil {
//		t.ErrEventHandler(retryDt.GetData(), err)
//	}
//}
//
//func (t *MiniTask) loopRetry() {
//	for {
//		select {
//		case val, ok := <-t.retryCh:
//			if !ok {
//				return
//			}
//			switch retryDt := val.(type) {
//			case RetryElem:
//				t.retryPush(retryDt)
//			default:
//				t.retryPush(t.newRetry(val))
//			}
//		}
//	}
//}
//
//func (t *MiniTask) do(data interface{}) {
//	switch val := data.(type) {
//	case RetryElem:
//		if t.exec(val.GetData()) {
//			t.retryPush(val)
//		}
//	default:
//		if t.exec(val) {
//			if err := t.Retry(val); err != nil {
//				t.ErrEventHandler(val, err)
//			}
//		}
//	}
//}
//
//func (t *MiniTask) loopExec(stopCh <-chan bool) {
//	ticker := time.NewTicker(time.Millisecond * 100)
//	for {
//		if t.isStop {
//			return
//		}
//		select {
//		case <-ticker.C:
//			data := t.dataList.Pop()
//			if data == nil {
//				break
//			}
//			t.do(data)
//		case <-stopCh:
//			return
//		}
//	}
//}
//
//func (t *MiniTask) clearRetryCh() {
//	close(t.retryCh)
//	var err error
//	var data interface{}
//	for val := range t.retryCh {
//		switch retryVal := val.(type) {
//		case RetryElem:
//			data = retryVal.GetData()
//		default:
//			data = retryVal
//		}
//		if err = t.pst.Store(data); err != nil {
//			t.ErrEventHandler(val, err)
//		}
//	}
//}
//
//func (t *MiniTask) clearQueue() {
//	var err error
//	for t.dataList.Length() > 0 {
//		data := t.dataList.Pop()
//		if data != nil {
//			// 清理数据的时候持久存储
//			if t.pst != nil {
//				if err = t.pst.Store(data); err != nil {
//					t.ErrEventHandler(data, err)
//				}
//			}
//		}
//	}
//}
//
//func (t *MiniTask) Push(ctx interface{}) error {
//	if err := t.dataList.Push(ctx); err != nil {
//		// 出现错误先放入重试
//		if err = t.Retry(ctx); err != nil {
//			return err
//		}
//	}
//	return nil
//}
//
//// 把数据放入到延迟队列中
//func (t *MiniTask) Retry(ctx interface{}) error {
//	if t.isStop {
//		return t.persistentStore(ctx)
//	}
//	if !t.RetryFlag {
//		return nil
//	}
//	select {
//	case t.retryCh <- ctx:
//	default:
//		return t.persistentStore(ctx)
//	}
//	return nil
//}
//
//func (t *MiniTask) delayPush(list []interface{}) {
//	// 如果延时队列到时间了就会自动发出要处理的数据
//	var err error
//	for _, val := range list {
//		if err = t.Push(val); err != nil {
//			task, ok := val.(RetryElem)
//			if ok {
//				t.ErrEventHandler(task.GetData(), err)
//			} else {
//				t.ErrEventHandler(val, err)
//			}
//		}
//	}
//}
//
//func (t *MiniTask) persistentStore(ctx interface{}) error {
//	if t.pst != nil {
//		return t.pst.Store(ctx)
//	}
//	return nil
//}
//
//func (t *MiniTask) persistentLoad() error {
//	if t.pst != nil {
//		list, err := t.pst.Load()
//		if err != nil {
//			return err
//		}
//		for i := 0; i < len(list); i++ {
//			err = t.dataList.Push(list[i])
//		}
//	}
//	return nil
//}
//
//func (t *MiniTask) autoScaleExec() {
//	tm := time.NewTicker(time.Second)
//	for {
//		select {
//		case <-tm.C:
//			// TODO
//		}
//	}
//}
