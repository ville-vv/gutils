package valid

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)
func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
func TestJwtToken_Generate(t *testing.T) {

	//signBytes, err := ioutil.ReadFile("./sample_key")
	//fatal(err)
	//
	//verifyBytes, err := ioutil.ReadFile("./sample_key.pub")
	//fatal(err)

	var pubBuff bytes.Buffer
	var prvBuff bytes.Buffer
	err := GenRsaKey(1024, &pubBuff, &prvBuff)
	fatal(err)

	//jwtT := JwtWithRsa{PubKey: verifyBytes, PrvKey:signBytes}
	jwtT := JwtWithRsa{PubKey: pubBuff.Bytes(), PrvKey:prvBuff.Bytes()}
	token, err := jwtT.Generate(map[string]string{"name":"cayla"}, 2)
	fmt.Println(token, err)
	fmt.Println(jwtT.Verify(token))
}