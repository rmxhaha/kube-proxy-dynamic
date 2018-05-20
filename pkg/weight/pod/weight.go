package pod

import (
	"sync"
	podloadstore "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/pod/store"
	"math"
)

type PodWeight struct {
	HostIP string
	PodIP string
	Weight uint16
}


type Store struct {
	m map[string]*PodWeight
	sync.Mutex
}

func (store *Store) Add(weight *PodWeight){
	store.Lock()
	defer store.Unlock()

	store.m[weight.PodIP] = weight
}

func (store *Store) GetWeight(ip string) uint16 {
	if pw, ok := store.m[ip]; !ok {
		return 0
	} else {
		return pw.Weight
	}
}

func convertLoadToWeight(load podloadstore.PodLoad) *PodWeight {
	return &PodWeight{ load.HostIP, load.PodIP, uint16(math.MaxUint16 - load.Load)}
}

func NewWeightStore(store *podloadstore.Store) (*Store) {
	podLoads := store.GetMap()

	wstore := &Store{ m: map[string]*PodWeight{}}
	for _, pl := range podLoads {
		wstore.Add(convertLoadToWeight(pl))
	}

	return wstore
}
