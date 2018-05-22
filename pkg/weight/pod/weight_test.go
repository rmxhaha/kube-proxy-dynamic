package pod

import (
	"testing"
	pls "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/pod/store"
	"fmt"
	"time"
)

func TestWeightProcessor_GetWeights(t *testing.T) {
	podloadstore := pls.New()

	collecttime := time.Now().Add( time.Duration(-1) * time.Second )

	podloadstore.Add(&pls.PodLoad{
		PodIP: "a",
		Load: 45000,
		RecordTime: collecttime,
	})

	podloadstore.Add(&pls.PodLoad{
		PodIP: "b",
		Load: 30000,
		RecordTime: collecttime,
	})

	podloadstore.Add(&pls.PodLoad{
		PodIP: "c",
		Load: 15000,
		RecordTime: collecttime,
	})


	wprocessor := NewWeightProcessor(podloadstore)

	weights := wprocessor.GetWeights([]string{"a","b","c"})

	for _, w := range weights {
		fmt.Println(w)
	}
}