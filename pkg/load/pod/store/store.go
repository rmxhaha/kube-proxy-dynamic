package store

import (
	"time"
	"sync"
)

type PodLoad struct {
	HostIP string
	PodIP string
	Load uint32
	RecordTime time.Time
}

type Store struct {
	m map[string]*PodLoad
	sync.Mutex
}

func New() *Store {
	return &Store{ m: map[string]*PodLoad {} }
}

func (store *Store) Add(podLoad *PodLoad){
	store.Lock()
	store.m[podLoad.PodIP] = podLoad
	store.Unlock()
}

func (store *Store) GetMap() map[string]PodLoad {
	m := map[string]PodLoad{}
	store.Lock()
	defer store.Unlock()
	for plk, pl := range store.m {
		m[plk] = *pl
	}

	return m
}