package rands

import (
	"fmt"
	"testing"
)

func TestGenVCode(t *testing.T) {
	for i := 0; i < 10000; i++ {
		fmt.Println(GenNumber(10))
	}
}

func TestRandLetterNumString(t *testing.T) {
	for i := 0; i < 10000; i++ {
		fmt.Println(GenLetterNumString(64))
	}
}
