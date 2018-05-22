package disk

import (
	"testing"
	"time"
	"fmt"
)

func TestDiskStatProvider_GetDiskUtilStat(t *testing.T) {
	p, err := NewDiskStatProvider(200*time.Millisecond)

	if err != nil {
		t.Fatalf("%s", err)
	}

	time.Sleep(300*time.Millisecond)

	dstat := p.GetDiskUtilStat()

	for name, s := range dstat {
		fmt.Printf("%s %f", name, s)
	}

}