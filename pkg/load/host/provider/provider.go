package provider

import (
	"errors"
	"time"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"github.com/rmxhaha/kube-proxy-dynamic/pkg/load/host/network"
	"github.com/rmxhaha/kube-proxy-dynamic/pkg/load/host/disk"
)

type Provider struct {
	networkRateProvider *network.NetworkRateProvider
	diskStatProvider *disk.DiskStatProvider
}

func New() (*Provider, error) {
	n := &Provider{}
	nrp, err := network.NewNetworkRateProvider(500 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	dsp, err := disk.NewDiskStatProvider(500 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	n.networkRateProvider = nrp
	n.diskStatProvider = dsp

	return n, nil
}

func (p *Provider) GetCPUUtil() (float64, error) {
	cpupercents, err := cpu.Percent(0, false)
	if err != nil {
		return 0, nil
	}

	return cpupercents[0]/100, nil
}

func (p *Provider) GetMemoryUtil() (float64, error) {
	memo, err := mem.VirtualMemory()

	if err != nil {
		return 0, nil
	}

	return memo.UsedPercent/100, nil
}

// id can be ipv4, ipv6, or interface name
func (p *Provider) GetNetworkUtil(id string) (float64, error){
	networkRateStats := p.networkRateProvider.GetNetworkRateStats()
	if v, ok := networkRateStats[id]; ok {
		return v.NaiveUsage/100, nil
	} else {
		return 0.0, errors.New("interface not found")
	}
}

func (p *Provider) GetNetworkMaxUtil() (float64, error){
	networkRateStats := p.networkRateProvider.GetNetworkRateStats()
	var maxUsage float64 = 0.0
	for _, stat := range networkRateStats {
		u := stat.NaiveUsage/100
		if maxUsage < u { maxUsage = u }
	}

	return maxUsage, nil
}



func (p *Provider) GetFSUtil() (float64, error) {
	stats := p.diskStatProvider.GetDiskUtilStat()
	maxutil := 0.0

	for _, util := range stats {
		if maxutil < util {
			maxutil = util
		}
	}

	return maxutil, nil
}
