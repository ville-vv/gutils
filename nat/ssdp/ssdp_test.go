package ssdp

import (
	"context"
	"fmt"
	"testing"
)

func TestSearch(t *testing.T) {
	serviceList, err := Search(context.Background(), SSDPAll)
	if err != nil {
		t.Fatal(err)
		return
	}

	for _, service := range serviceList {
		fmt.Println(service.ST)
	}

}
