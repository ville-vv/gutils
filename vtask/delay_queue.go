package vtask

//
//type delayTaskFunc func(interface{})
//
//type delayTask struct {
//	cycleNum int
//	param    interface{}
//	exec     delayTaskFunc
//}
//
//func (sel *delayTask) Dec() {
//	sel.cycleNum--
//}
//
//func (sel *delayTask) IsExpire() bool {
//	return sel.cycleNum <= 0
//}
//
//type DelayTask struct {
//	slots       [60]*sync.Map
//	slotTaskNum [60]*AtomicInt64
//	curIndex    int
//	maxTaskSize int64
//	closed      bool
//	stopCh      chan bool
//	timeTick    *time.Ticker
//	taskCh      chan delayTask
//}
//
//func NewDelayTask(size ...int64) *DelayTask {
//	if len(size) == 0 {
//		size = append(size, 1000)
//	}
//	dt := &DelayTask{
//		slots:       [60]*sync.Map{},
//		slotTaskNum: [60]*AtomicInt64{},
//		curIndex:    0,
//		maxTaskSize: size[0],
//		stopCh:      make(chan bool),
//		timeTick:    time.NewTicker(time.Second),
//		taskCh:      make(chan delayTask, size[0]),
//	}
//
//	for i := 0; i < 60; i++ {
//		dt.slots[i] = &sync.Map{}
//	}
//
//	for i := 0; i < 60; i++ {
//		dt.slotTaskNum[i] = &AtomicInt64{}
//	}
//
//	go dt.loopExec()
//	go dt.loopTask()
//
//	return dt
//}
//
//func (sel *DelayTask) Close() {
//	sel.closed = true
//	close(sel.stopCh)
//	close(sel.taskCh)
//}
//
//func (sel *DelayTask) loopExec() {
//	for {
//		select {
//		case <-sel.stopCh:
//			return
//		case task, ok := <-sel.taskCh:
//			if !ok {
//				return
//			}
//			go task.exec(task.param)
//		}
//	}
//}
//
//func (sel *DelayTask) next() {
//	sel.curIndex++
//	if sel.curIndex >= 60 {
//		sel.curIndex = 0
//	}
//}
//
//func (sel *DelayTask) check() {
//	tasks := sel.slots[sel.curIndex]
//	taskNum := sel.slotTaskNum[sel.curIndex]
//	tasks.Range(func(key, value interface{}) bool {
//		task, _ := value.(*delayTask)
//		if task.IsExpire() {
//			select {
//			case sel.taskCh <- *task:
//				tasks.Delete(key)
//				taskNum.Dec()
//			}
//		} else {
//			// 没有到期就轮询值减一个
//			task.Dec()
//		}
//		return true
//	})
//}
//
//func (sel *DelayTask) loopTask() {
//	for {
//		select {
//		case <-sel.stopCh:
//			sel.timeTick.Stop()
//			return
//		case <-sel.timeTick.C:
//			sel.check()
//			sel.next()
//		}
//	}
//}
//
//func (sel *DelayTask) slotIdx(subSecond int) int {
//	idx := (subSecond)%60 + sel.curIndex
//	if idx >= 60 {
//		idx = idx - 60
//	}
//	return idx
//}
//
//func (sel *DelayTask) Push(name string, params interface{}, taskF delayTaskFunc, tm time.Time) error {
//	if sel.closed {
//		return nil
//	}
//	timeNow := time.Now()
//	if tm.Before(timeNow) {
//		return errors.New("time must be after than now")
//	}
//	subSecond := int(tm.Unix() - timeNow.Unix())
//	idx := sel.slotIdx(subSecond)
//	cycleNum := subSecond / 60
//	if sel.slotTaskNum[idx].Load() > sel.maxTaskSize {
//		return nil
//	}
//	sel.slotTaskNum[idx].Inc()
//	sel.slots[idx].Store(name, &delayTask{
//		cycleNum: cycleNum,
//		param:    params,
//		exec:     taskF,
//	})
//	return nil
//}
//
//type delayNode struct {
//	cycleNum int
//	value    interface{}
//}
//
//func (sel *delayNode) IsExpire() bool {
//	return sel.cycleNum <= 0
//}
//func (sel *delayNode) Dec() {
//	sel.cycleNum--
//}
//
//type DelayQueue struct {
//	slots       [60]*sync.Map
//	slotLen     [60]*AtomicInt64
//	tempList    sync.Pool
//	curIndex    int
//	maxTaskSize int64
//	started     bool
//	stopCh      chan bool
//	timeTick    *time.Ticker
//	taskCh      chan delayNode
//	workFunc    func([]interface{})
//	once        sync.Once
//}
//
//func NewDelayQueue(workFunc func([]interface{}), size ...int64) *DelayQueue {
//	if len(size) == 0 {
//		size = append(size, 100000)
//	}
//	dt := &DelayQueue{
//		slots:       [60]*sync.Map{},
//		slotLen:     [60]*AtomicInt64{},
//		curIndex:    0,
//		maxTaskSize: size[0],
//		stopCh:      make(chan bool),
//		timeTick:    time.NewTicker(time.Second),
//		taskCh:      make(chan delayNode, size[0]),
//		workFunc:    workFunc,
//		tempList: sync.Pool{New: func() interface{} {
//			return make([]interface{}, 0, size[0])
//		},
//		},
//	}
//	for i := 0; i < 60; i++ {
//		dt.slots[i] = &sync.Map{}
//	}
//	for i := 0; i < 60; i++ {
//		dt.slotLen[i] = &AtomicInt64{}
//	}
//	return dt
//}
//
//func (sel *DelayQueue) Run() {
//	sel.once.Do(func() {
//		sel.started = true
//		go sel.loopTask()
//	})
//}
//
//func (sel *DelayQueue) Close() {
//	sel.started = false
//	close(sel.stopCh)
//	close(sel.taskCh)
//}
//
//func (sel *DelayQueue) next() {
//	sel.curIndex++
//	if sel.curIndex >= 60 {
//		sel.curIndex = 0
//	}
//}
//
//func (sel *DelayQueue) check() {
//	tasks := sel.slots[sel.curIndex]
//	taskNum := sel.slotLen[sel.curIndex]
//	tempList := sel.tempList.Get().([]interface{})
//	tasks.Range(func(key, value interface{}) bool {
//		task, _ := value.(*delayNode)
//		if task.IsExpire() {
//			tasks.Delete(key)
//			taskNum.Dec()
//			tempList = append(tempList, task.value)
//		} else {
//			task.Dec()
//		}
//		return true
//	})
//	if len(tempList) > 0 {
//		sel.workFunc(tempList)
//		tempList = nil
//	}
//	sel.tempList.Put(tempList)
//}
//
//func (sel *DelayQueue) loopTask() {
//	for {
//		select {
//		case <-sel.stopCh:
//			sel.timeTick.Stop()
//			return
//		case <-sel.timeTick.C:
//			sel.check()
//			sel.next()
//		}
//	}
//}
//
//func (sel *DelayQueue) slotIdx(subSecond int) int {
//	idx := (subSecond)%60 + sel.curIndex
//	if idx >= 60 {
//		idx = idx - 60
//	}
//	return idx
//}
//
//// Push
//// val 要加入队列的值
//// tm 延迟时间最小单位为秒
//func (sel *DelayQueue) Push(val interface{}, tm int64) error {
//	return sel.push(rands.GenLetterString(16), val, int(tm))
//}
//
//func (sel *DelayQueue) push(name string, val interface{}, subSecond int) error {
//	if !sel.started {
//		return ErrDelayNotStarted
//	}
//	if subSecond <= 0 {
//		subSecond = 1
//	}
//	idx := sel.slotIdx(subSecond)
//	if sel.slotLen[idx].Load() > sel.maxTaskSize {
//		return ErrOverMaxSize
//	}
//	sel.slotLen[idx].Inc()
//	sel.slots[idx].Store(name, &delayNode{
//		cycleNum: subSecond / 60,
//		value:    val,
//	})
//	return nil
//}
//
//type delayTaskV2 struct {
//	Name  string
//	param interface{}
//	exec  delayTaskFunc
//}
//
//type DelayTaskV2 struct {
//	delayQueue *DelayQueue
//}
//
//func NewDelayTaskV2() *DelayTaskV2 {
//	tsk := &DelayTaskV2{}
//	return tsk
//}
//
//func (sel *DelayTaskV2) Start() {
//	sel.delayQueue = NewDelayQueue(sel.exec)
//	sel.delayQueue.Run()
//}
//
//func (sel *DelayTaskV2) Close() {
//	sel.delayQueue.Close()
//}
//
//func (sel *DelayTaskV2) exec(list []interface{}) {
//	l := len(list)
//	for i := 0; i < l; i++ {
//		task, ok := list[i].(*delayTaskV2)
//		if !ok {
//			continue
//		}
//		task.exec(task.param)
//	}
//}
//
//func (sel *DelayTaskV2) Push(name string, params interface{}, tm int64, taskF delayTaskFunc) error {
//	return sel.delayQueue.Push(&delayTaskV2{
//		Name:  name,
//		param: params,
//		exec:  taskF,
//	}, tm)
//}
