package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestIDgen(t *testing.T) {
	for i := 0; i < 10; i++ {
		a := GetID()
		<-time.After(time.Millisecond * 100)
		fmt.Printf("id %d is : %s , len is : %d\n", i, a, len(a))
	}
}
