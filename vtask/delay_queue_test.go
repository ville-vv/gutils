package vtask

//func TestDelayQueue_Loop(t *testing.T) {
//	dqu := NewDelayTask()
//
//	taskF := func(p interface{}) {
//		//fmt.Println(p)
//	}
//
//	//time.Sleep(time.Second * 5)
//
//	for i := 0; i < 10; i++ {
//		go func(n int) {
//			for {
//				time.Sleep(time.Second * 1)
//				name := rands.GenLetterString(10)
//				_ = dqu.Push(name, []string{name}, taskF, time.Now().Add(time.Second*80))
//			}
//
//		}(i)
//	}
//
//	time.Sleep(time.Second * 5)
//
//	dqu.Close()
//	time.Sleep(time.Second)
//}
//
//func TestConsterList(t *testing.T) {
//	lt := list.New()
//	for i := 0; i < 10; i++ {
//		lt.PushFront(i)
//	}
//
//	for i := 0; i < 10; i++ {
//		fmt.Println(lt.Back())
//	}
//
//}
//
//func TestDelayQueue_Push(t *testing.T) {
//
//	addTimeCnt := AtomicInt64{}
//	timesCnt := 0
//
//	taskF := func(pList []interface{}) {
//		timesCnt += len(pList)
//		fmt.Println(addTimeCnt.Load(), timesCnt)
//	}
//	dqu := NewDelayQueue(taskF)
//	dqu.Run()
//	for i := 0; i < 100; i++ {
//		go func(n int) {
//			for {
//				time.Sleep(time.Second * 1)
//				name := rands.GenLetterString(16)
//				_ = dqu.Push(name, 1)
//				addTimeCnt.Inc()
//				//return
//				if addTimeCnt.Load() > 1000 {
//					return
//				}
//			}
//		}(i)
//	}
//	time.Sleep(time.Second * 20)
//	dqu.Close()
//}
//
//func BenchmarkDelayQueue_Push(b *testing.B) {
//	b.StopTimer()
//	taskF := func(pList []interface{}) {
//	}
//	dqu := NewDelayQueue(taskF)
//	dqu.Run()
//	b.StartTimer()
//	for i := 0; i < b.N; i++ {
//		// BenchmarkDelayQueue_Push-4   	10580889	       110 ns/op
//		dqu.Push(rands.GenLetterString(16), 1)
//	}
//	dqu.Close()
//}
