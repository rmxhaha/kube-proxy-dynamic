package collector

import (
	"google.golang.org/grpc"
	"time"
	"log"
	"k8s.io/client-go/kubernetes"
	"fmt"
	"k8s.io/client-go/tools/cache"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/api/core/v1"
	informerv1 "k8s.io/client-go/informers/core/v1"

	"github.com/rmxhaha/kube-proxy-dynamic/pkg/load/pod/store"
)

type Collector struct {
	nodeInformer cache.SharedIndexInformer
	workers      map[string]*worker
	podLoadStore *store.Store
	dialOpts []grpc.DialOption
	dialPort int
}

func New(store *store.Store, clientset kubernetes.Interface, dialPort int, dialOpts []grpc.DialOption) (*Collector, error) {
	lc := &Collector{}

	lc.dialOpts = dialOpts
	lc.dialPort = dialPort

	nodeInformer := informerv1.NewNodeInformer(clientset, 5 * time.Second, cache.Indexers{})
	lc.nodeInformer = nodeInformer

	nodeInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    lc.addNode,
		UpdateFunc: lc.updateNode,
		DeleteFunc: lc.removeNode,
	})

	lc.workers = map[string]*worker{}
	lc.podLoadStore = store

	return lc, nil
}

func (lc *Collector) Run() {
	lc.nodeInformer.Run(wait.NeverStop)
}

func (lc *Collector) addNode(n interface{}){
	node := n.(*v1.Node)

	log.Println(node.Name)

	addr := ""

	for _, a := range node.Status.Addresses {
		if a.Type == v1.NodeInternalIP {
			addr = a.Address
		}
	}

	lcw, err := newWorker(fmt.Sprintf("%s:%d", addr, lc.dialPort), lc.podLoadStore)
	if err != nil {
		log.Println(err)
		return
	}

	lc.workers[node.Name] = lcw

	go lcw.Run()
}

func (lc *Collector) updateNode(o interface{}, n interface{}){
}

func (lc *Collector) removeNode(n interface{}){
	node := n.(*v1.Node)

	lc.workers[node.Name].ShouldQuit = true

	delete(lc.workers, node.Name)
}
