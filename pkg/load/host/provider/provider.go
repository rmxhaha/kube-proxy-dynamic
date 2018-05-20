package provider

import (
	"errors"
	"time"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/disk"
	"github.com/rmxhaha/kube-proxy-dynamic/pkg/load/host/network"
)

type Provider struct {
	networkRateProvider *network.NetworkRateProvider
}

func New() (*Provider, error) {
	n := &Provider{}
	nrp, err := network.NewNetworkRateProvider(500 * time.Millisecond)
	if err != nil {
		return nil, err
	}
	n.networkRateProvider = nrp

	return n, nil
}

func (p *Provider) GetCPUUsage() (float64, error) {
	cpupercents, err := cpu.Percent(0, false)
	if err != nil {
		return 0, nil
	}

	return cpupercents[0]/100, nil
}

func (p *Provider) GetMemoryUsage() (float64, error) {
	memo, err := mem.VirtualMemory()

	if err != nil {
		return 0, nil
	}

	return memo.UsedPercent/100, nil
}

// id can be ipv4, ipv6, or interface name
func (p *Provider) GetNetworkUsage(id string) (float64, error){
	networkRateStats := p.networkRateProvider.GetNetworkRateStats()
	if v, ok := networkRateStats[id]; ok {
		return v.NaiveUsage/100, nil
	} else {
		return 0.0, errors.New("interface not found")
	}
}

func (p *Provider) GetNetworkMaxUsage() (float64, error){
	networkRateStats := p.networkRateProvider.GetNetworkRateStats()
	var maxUsage float64 = 0.0
	for _, stat := range networkRateStats {
		u := stat.NaiveUsage/100
		if maxUsage < u { maxUsage = u }
	}

	return maxUsage, nil
}

func (p *Provider) GetFSUsage() (float64, error) {
	disk.Usage()

}
