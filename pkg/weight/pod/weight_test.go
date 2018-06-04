package pod

import (
	"testing"
	pls "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/pod/store"
	"fmt"
	"time"
	"math"
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


	wprocessor := NewWeightProcessor(podloadstore, uint16(10))

	weights := wprocessor.GetWeights([]string{"a","b","c"})

	for _, w := range weights {
		fmt.Println(w)
	}
}

func TestWeightProcessor_GetWeights2(t *testing.T) {
	// percent utilization A,B,C,  Information Age(ms),  Weight output
	tests := [][]int {
		{100,100,100, 1000,   		10, 10, 10 },
		{100,100,100, 500,     		10, 10, 10 },
		{100,100,100, 250,     		10, 10, 10 },
		{100,100,100, 1000000,     	10, 10, 10 },
		{0,0,0		, 1000, 		10, 10, 10 },
		{0,0,0		, 500, 			10, 10, 10 },
		{0,0,0		, 250, 			10, 10, 10 },
		{0,50,100	, 250, 			10, 2, 0 },
		{0,50,100	, 500, 			10, 5, 0 },
		{0,50,100	, 1000, 		10, 6, 3 },
		{0,50,100	, 2000, 		10, 8, 5 },
	}

	for tcno, test := range tests {
		aload := uint32(math.MaxUint16*test[0]/100)
		bload := uint32(math.MaxUint16*test[1]/100)
		cload := uint32(math.MaxUint16*test[2]/100)

		aout := uint8(test[4])
		bout := uint8(test[5])
		cout := uint8(test[6])

		now := time.Now()
		collecttime := now.Add( time.Duration(-test[3]) * time.Millisecond )


		weightrange := uint16(10)
		podloadstore := pls.New()

		podloadstore.Add(&pls.PodLoad{
			PodIP: "a",
			Load: aload,
			RecordTime: collecttime,
		})

		podloadstore.Add(&pls.PodLoad{
			PodIP: "b",
			Load: bload,
			RecordTime: collecttime,
		})

		podloadstore.Add(&pls.PodLoad{
			PodIP: "c",
			Load: cload,
			RecordTime: collecttime,
		})


		wprocessor := NewWeightProcessor(podloadstore, weightrange)

		weights := wprocessor.getweights([]string{"a","b","c"}, now)

		if weights["a"] != aout+1 {
			t.Fatalf("TC %d expected %d: Got %d", tcno, aout, weights["a"]-1)
		}

		if weights["b"] != bout+1 {
			t.Fatalf("TC %d expected %d: Got %d", tcno, bout, weights["b"]-1)
		}

		if weights["c"] != cout+1 {
			t.Fatalf("TC %d expected %d: Got %d", tcno, cout, weights["c"]-1)
		}

	}


}

