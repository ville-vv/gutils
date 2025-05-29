package bytepack

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRegister(t *testing.T) {
	p := NewPack()

	buf := bytes.NewBufferString("")
	err := p.WriteMsg(buf, "Hello Yes")
	assert.NoError(t, err)
	fmt.Println(p.ReadMsg(buf))
	//fmt.Println(p.ReadMsg(buf))
}
