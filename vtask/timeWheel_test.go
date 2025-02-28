package vtask

import (
	"fmt"
	"testing"
	"time"
)

func TestNewTimeWheel(t *testing.T) {
	tw := NewTimeWheel(WithTimeWheelInterval(time.Millisecond*10), WithTimeWheelSlotsNum(600))
	if tw == nil {
		t.Fatal("NewTimeWheel failed")
	}
	tw.Start()

	addFn := func() {
		for i := 0; i < 10000; i++ {
			tw.AddTask(time.Millisecond*10, func(param interface{}) {
				fmt.Println("task:", param, tw.Len())
			}, i)
		}
	}

	for i := 0; i < 10999; i++ {
		go addFn()
	}

	time.Sleep(time.Second * 10)

	tw.Stop()

}
