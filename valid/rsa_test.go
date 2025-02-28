package valid

import (
	"bytes"
	"fmt"
	"testing"
)

func TestGenRsaKey(t *testing.T) {
	var pubBuff bytes.Buffer
	var prvBuff bytes.Buffer

	err := GenRsaKey(1024, &pubBuff, &prvBuff)
	t.Log(err)

	signData := RsaSignWithSha256([]byte("hello I am Chinese"), prvBuff.Bytes())
	fmt.Println(RsaVerySignWithSha256([]byte("hello I am Chinese"), signData, pubBuff.Bytes()))

}
