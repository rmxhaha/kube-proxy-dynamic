package provider

import "time"
import (
	"k8s.io/client-go/tools/cache"
	"math"
	"fmt"
	"github.com/golang/protobuf/ptypes"
	pb "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/exchange/loadexchange"
	corev1 "k8s.io/api/core/v1"
	"os"
	informerv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	hostprovider "github.com/rmxhaha/kube-proxy-dynamic/pkg/load/host/provider"
	"github.com/rmxhaha/kube-proxy-dynamic/pkg/load/host/kubeletsummary"
)

type Provider struct {
	hostLoadProvider *hostprovider.Provider
	HostIP string
	podInformer  cache.SharedIndexInformer
	kubeClientSet *kubernetes.Clientset
}

func (plp *Provider) initHostLoadProvider() error {
	hostLoadProvider, err := hostprovider.New()
	if err != nil {
		return err
	}
	plp.hostLoadProvider = hostLoadProvider
	return nil
}

func (plp *Provider) initHostIP() error {

	hostname, err := os.Hostname()

	node, err := plp.kubeClientSet.CoreV1().Nodes().Get(hostname, metav1.GetOptions{})
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to retrieve node info: %v", err))
	}
	address := node.Status.Addresses[0].Address
	plp.HostIP = address
	return nil
}

func (plp *Provider) initInformers() error {
	podInformer := informerv1.NewPodInformer(plp.kubeClientSet, "", 5 * time.Second, cache.Indexers{})

	plp.podInformer = podInformer
	go podInformer.Run(wait.NeverStop)

	return nil
}


func New(kubeClientSet *kubernetes.Clientset) (*Provider, error) {
	plp := &Provider{
		kubeClientSet:kubeClientSet,
	}
	if err := plp.initHostLoadProvider(); err != nil {
		return nil, err
	}
	if err := plp.initHostIP(); err != nil {
		return nil, err
	}
	if err := plp.initInformers(); err != nil {
		return nil, err
	}
	return plp, nil
}

func (plp *Provider) GetPodLoads() (*pb.PodLoads, error) {
	podLoads := &pb.PodLoads{}

	var hostPercent int64 = -1
	var hostCPUPercent int64 = 0
	var hostMemoryPercent int64 = 0
	var hostNetworkPercent int64 = 0
	var hostFSPercent int64 = 0

	hostcpuval, err := plp.hostLoadProvider.GetCPUUsage()
	if err != nil {
		fmt.Println(err)
	} else {
		hostCPUPercent = int64(hostcpuval * math.MaxUint16)
		if hostPercent < hostCPUPercent { hostPercent = hostCPUPercent}
	}

	hostmemoryval, err := plp.hostLoadProvider.GetMemoryUsage()
	if err != nil {
		fmt.Println(err)
	} else {
		hostMemoryPercent = int64(hostmemoryval * math.MaxUint16)
		if hostPercent < hostMemoryPercent { hostPercent = hostMemoryPercent}
	}

	hostnetworkval, err := plp.hostLoadProvider.GetNetworkUsage(plp.HostIP)
	if err != nil {
		fmt.Println(err)
	} else {
		hostNetworkPercent = int64(hostnetworkval * math.MaxUint16)
		if hostPercent < hostNetworkPercent { hostPercent = hostNetworkPercent}
	}

	hostfsval, err := plp.hostLoadProvider.GetFSUsage()
	if err != nil {
		fmt.Println(err)
	} else {
		hostFSPercent = int64(hostfsval * math.MaxUint16)
		if hostPercent < hostFSPercent { hostPercent = hostFSPercent }
	}

	summary, err := kubeletsummary.GetLocalSummary()
	if err != nil {
		return nil, err
	}

	cpuloads := map[string]int64 {}
	memoryloads := map[string]int64 {}

	for _, p := range summary.Pods {
		for _, c := range p.Containers {
			if c.CPU.UsageNanoCores != nil && c.Memory.UsageBytes != nil {
				cpuloads[p.PodRef.UID+c.Name] = int64(*c.CPU.UsageNanoCores) / 1000000
				memoryloads[p.PodRef.UID+c.Name] = int64(*c.Memory.UsageBytes) / 1000000
			}
		}
	}


	for _, p := range plp.podInformer.GetStore().List() {
		pod := p.(*corev1.Pod)
		if pod.Status.HostIP != plp.HostIP {
			continue
		}
		if pod.Status.PodIP == "" {
			continue
		}

		var maxPercent int64 = -1
		for _,c := range pod.Spec.Containers {
			key := string(pod.UID) + c.Name

			var CPUPercent int64 = 0
			var memoryPercent int64 = 0

			cpuval, cpuok := cpuloads[key]
			memoval, memook := memoryloads[key]

			if cpuok {
				if c.Resources.Limits.Cpu().IsZero() { // assume host load because there is no limit the host is the limit
					if maxPercent < hostCPUPercent {
						maxPercent = hostCPUPercent
					}
				} else {
					CPUPercent = cpuval * int64(math.MaxUint16) / c.Resources.Limits.Cpu().ScaledValue(-3)
					if maxPercent < CPUPercent {
						maxPercent = CPUPercent
					}
				}
			}

			if memook {
				if c.Resources.Limits.Memory().IsZero() {
					if maxPercent < hostMemoryPercent {
						maxPercent = hostMemoryPercent
					}
				} else {
					memoryPercent = memoval * int64(math.MaxUint16) / c.Resources.Limits.Memory().ScaledValue(6)
					if maxPercent < memoryPercent {
						maxPercent = memoryPercent
					}
				}
			}

			if maxPercent < hostNetworkPercent {
				maxPercent = hostNetworkPercent
			}

			if maxPercent < hostFSPercent {
				maxPercent = hostFSPercent
			}
		}

		// assume metrics not yet ready.
		if maxPercent == -1 {
			continue
		}

		ip := net.ParseIP(pod.Status.PodIP).To4()

		podLoad := &pb.PodLoad{PodIP: []byte(ip), Load: uint32(maxPercent)}
		fmt.Println(pod.Name, pod.Status.PodIP, maxPercent)
		podLoads.PodLoads = append(podLoads.PodLoads, podLoad)
	}

	recordTime, err := ptypes.TimestampProto(time.Now())

	if err == nil {
		podLoads.RecordTime = recordTime
	}

	return podLoads, nil
}