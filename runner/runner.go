package runner

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
)

func PrintPanicStack(fn func(s string), extras ...interface{}) {
	if err := recover(); err != nil {
		var stack strings.Builder
		if len(extras) > 0 {
			stack.WriteString(fmt.Sprintf("recover: %v %v\n", err, extras))
		} else {
			stack.WriteString(fmt.Sprintf("recover: %v\n", err))
		}
		// 打印调用栈信息
		for i := 0; ; i++ {
			funcName, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			stack.WriteString(fmt.Sprintf("%d: file: %s %d, %s\n", i, filepath.Base(file), line, runtime.FuncForPC(funcName).Name()))
		}
		fn(stack.String())
	}
}

func GoSafe(fn func()) {
	go func() {
		defer PrintPanicStack(func(s string) {
			fmt.Printf(s)
		})
	}()
}
