package vtask

//
//func TestChan(t *testing.T) {
//
//	testCh := make(chan int, 100)
//	for i := 0; i < 100; i++ {
//		testCh <- i
//	}
//	close(testCh)
//	go func() {
//		time.Sleep(time.Millisecond * 2)
//		for {
//			select {
//			case val, ok := <-testCh:
//				if !ok {
//					return
//				}
//				fmt.Println(val)
//			}
//		}
//	}()
//	time.Sleep(time.Millisecond)
//}
//
//type RetryElemMock struct {
//	data    interface{}
//	timesDo int
//}
//
//func (s *RetryElemMock) Can() bool {
//	if s.timesDo <= 0 {
//		return false
//	}
//	s.timesDo--
//	return true
//}
//
//func (s *RetryElemMock) Interval() int64 {
//	return 4
//}
//
//func (s *RetryElemMock) GetData() interface{} {
//	return s.data
//}
//
//type PersistentMock struct {
//}
//
//func (p *PersistentMock) Load() ([]interface{}, error) {
//	return nil, nil
//}
//
//func (p *PersistentMock) Store(i interface{}) error {
//	fmt.Println("持久化存储：", i)
//	return nil
//}
//
//func TestNewMiniTask(t *testing.T) {
//	miniTask := NewMiniTask(&TaskOption{
//		RetryFlag:  true,
//		Persistent: &PersistentMock{},
//		NewRetry: func(val interface{}) RetryElem {
//			return &RetryElemMock{
//				data:    val,
//				timesDo: 3,
//			}
//		},
//		ErrEventHandler: func(ctx interface{}, err error) {
//			fmt.Println(err)
//			return
//		},
//		Exec: func(val interface{}) (retry bool) {
//			fmt.Println("当前数据：", val)
//			return true
//		},
//	})
//	miniTask.Start()
//
//	for i := 0; i < 100; i++ {
//		go func(n int) {
//			for {
//				time.Sleep(time.Second * 1)
//				miniTask.Push(fmt.Sprintf("%d", n))
//				return
//			}
//		}(i)
//	}
//
//	//select {}
//
//	time.Sleep(10 * time.Second)
//
//	miniTask.Stop()
//
//}
