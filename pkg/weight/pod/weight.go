package pod

import (
	podloadstore "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/pod/store"
	"math"
	"sort"
	"time"
)

type WeightProcessor struct {
	podloadstore *podloadstore.Store
	weightrange uint16
}

func (processor *WeightProcessor) GetWeights(ips []string) map[string]uint8 {
	return processor.getweights(ips, time.Now())
}


func (processor *WeightProcessor) getweights(ips []string, now time.Time) map[string]uint8 {
	m := processor.podloadstore.GetMap()
	weights := map[string]uint8 {}

	podloads := make([]podloadstore.PodLoad,0)

	for _, ip := range ips {
		if val, ok := m[ip]; ok {
			podloads = append(podloads, val)
		}
	}

	if len(podloads) == 0 {
		return weights
	}

	sort.Slice( podloads, func(i,j int) bool {
		return podloads[i].Load < podloads[j].Load
	})


	var sumagenano int64 = 0
	for _, pl := range podloads {
		d := now.Sub(pl.RecordTime)
		sumagenano += d.Nanoseconds()
	}

	averageage := float64(sumagenano / int64(len(podloads))) / 1000000000.0

	// expected capacity used in capacity in averageage
	arrivet := averageage * float64(len(podloads))

	sumcapacity := func (l int) float64 {
		var total float64 = 0
		for i := 0; i < l; i ++ {
			capacity := float64(podloads[l-1].Load - podloads[i].Load) / float64(math.MaxUint16)
			total += capacity
		}
		return total
	}

	l := len(podloads)
	capacity := 0.0
	for  l >= 0 {
		capacity = sumcapacity(l)
		if capacity < arrivet {
			break
		}
		l --
	}

	if l <= 0 { return weights }

	excess := (arrivet - capacity) / float64(l)

	fweights := make([]float64, l)
	maxfweight := 0.000001
	for i := 0; i < l; i ++ {
		capacity := float64(podloads[l-1].Load - podloads[i].Load) / float64(math.MaxUint16)
		fweight := capacity + excess
		if maxfweight < fweight { maxfweight = fweight }

		fweights[i] = fweight
	}

	for i := 0; i < l; i ++ {
		weights[podloads[i].PodIP] = uint8( 1 + fweights[i] / maxfweight * float64(processor.weightrange) )
	}

	for i := l; i < len(podloads); i++ {
		weights[podloads[i].PodIP] = uint8(1) // default to one if not
	}

	return weights
}


func NewWeightProcessor(store *podloadstore.Store, weightrange uint16) (*WeightProcessor) {
	wstore := &WeightProcessor{ podloadstore: store, weightrange: weightrange }

	return wstore
}
