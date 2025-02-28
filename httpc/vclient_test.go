package httpc

import (
	"fmt"
	"testing"
)

func TestVClient_Get(t *testing.T) {
	reqBody, err := Do().Get("http://www.baidu.com", nil)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Printf("%s", string(reqBody))
}
