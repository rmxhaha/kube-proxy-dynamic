package server

import (
	"sync"
	"time"
	pb "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/exchange/loadexchange"
	"k8s.io/client-go/kubernetes"
	"fmt"
	podloadprovider "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/pod/provider"
)

type worker struct {
	Source chan *pb.PodLoads
	Quit chan struct {}
}

func newWorker() *worker {
	return &worker{ Source: make(chan *pb.PodLoads, 4) }
}


type Server struct {
	workerMutex *sync.Mutex
	workers []*worker
	updateInterval time.Duration
	podLoadProvider *podloadprovider.Provider
	kubeClientSet *kubernetes.Clientset
}

func New(kubeClientSet *kubernetes.Clientset, updateInterval time.Duration) (*Server, error) {
	plp, err := podloadprovider.New(kubeClientSet)
	if err != nil { return nil, err }

	return &Server{
		podLoadProvider: plp,
		workerMutex: &sync.Mutex{},
		updateInterval: updateInterval,
		kubeClientSet:kubeClientSet,
	}, nil
}

func (les *Server) Run() {
	for {
		les.update()
		time.Sleep(les.updateInterval)
	}
}

func (les*Server) GetPodLoads(selector *pb.PodSelector, stream pb.LoadExchange_GetPodLoadsServer) error {
	worker := newWorker()

	wid := len(les.workers)

	les.workerMutex.Lock()
	les.workers = append(les.workers, worker)
	les.workerMutex.Unlock()

	defer func(){
		for k, w := range les.workers {
			if w == worker {
				fmt.Println("deleted")
				les.workers = append(les.workers[:k],les.workers[k+1:]...)
			}
		}
	}()

	for {
		select {
		case msg := <- worker.Source:
			fmt.Printf("send %d\n", wid)
			err := stream.Send(msg)
			if err != nil {
				return err
			}
		case <-worker.Quit:
			break
		}
	}

	return nil
}


func (les *Server) update() error {
	podLoads, err := les.podLoadProvider.GetPodLoads()

	if err != nil { return err }

	for _, w := range les.workers {
		w.Source <- podLoads
	}

	return nil
}