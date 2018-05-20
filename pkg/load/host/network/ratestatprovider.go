package network

import (
	"time"
	utilnet "github.com/shirou/gopsutil/net"

	"github.com/pkg/errors"
)

type NetworkRateProvider struct {
	lastStat              []utilnet.IOCountersStat
	lastCollect           time.Time
	updateInterval        time.Duration
	latestNetworkRateStat NetworkRateStats
	networkSpec           *NetworkSpec
}


func NewNetworkRateProvider(updateInterval time.Duration) (*NetworkRateProvider, error){
	n := &NetworkRateProvider{}

	if err := n.populateStat(); err != nil {
		return nil, err
	}

	if err := n.populateSpec(); err != nil {
		return nil, err
	}

	n.updateInterval = updateInterval
	go func(){
		for {
			time.Sleep(n.updateInterval)
			n.updateStat() // ignore error
		}
	}()

	return n, nil
}

func (n *NetworkRateProvider) populateSpec() error {
	networkSpec, err := NewHostNetworkSpec()
	if err != nil {
		return err
	}

	n.networkSpec = networkSpec
	return nil
}

func (n *NetworkRateProvider) populateStat() error {
	stat, err := utilnet.IOCounters(true)
	if err != nil {
		return err
	}

	n.lastStat = stat
	n.lastCollect = time.Now()
	return nil
}

func (n *NetworkRateProvider) updateStat() error {
	stats1 := n.lastStat
	previousCollectTime := n.lastCollect

	statsmap1 := map[string]utilnet.IOCountersStat {}
	for _, s := range stats1 {
		statsmap1[s.Name] = s
	}

	if err := n.populateStat(); err != nil {
		return err
	}

	stats2 := n.lastStat
	currentCollectTime := n.lastCollect

	deltaTimeMilli := uint64(currentCollectTime.Sub(previousCollectTime).Nanoseconds()) / 1000000


	if len(stats1) != len(stats2) {
		return errors.New("cannot handle new or deleted network interface")
	}

	rateStats := NetworkRateStats {}

	for _, s2 := range stats2 {
		s1 := statsmap1[s2.Name]
		rateStat := &NetworkRateStat{}
		rateStat.Name = s2.Name
		rateStat.BytesRecvRate = (s2.BytesRecv - s1.BytesRecv) / deltaTimeMilli * 1000
		rateStat.BytesSentRate = (s2.BytesSent - s1.BytesSent) / deltaTimeMilli * 1000
		rateStat.PacketsRecvRate = (s2.PacketsRecv - s1.PacketsRecv) / deltaTimeMilli * 1000
		rateStat.PacketsSentRate = (s2.PacketsSent - s1.PacketsSent) / deltaTimeMilli * 1000
		rateStat.DateRecorded = currentCollectTime
		OneMbps := int64(100000000 / 8)
		if iLimit := n.networkSpec.GetSpeed(s2.Name); iLimit != nil {
			if iLimit.Value() < OneMbps {
				// assume bug and defaults to 100MBit
				rateStat.BytesSpeed = uint64(OneMbps)
			} else {
				rateStat.BytesSpeed = uint64(iLimit.ScaledValue(0))
			}
		} else {
			// network device not found assume 1Mbps
			rateStat.BytesSpeed = uint64(OneMbps)
		}

		rateStat.NaiveUsage = float64(rateStat.BytesSentRate+rateStat.BytesRecvRate) / float64(rateStat.BytesSpeed) * 100

		rateStats[s2.Name] = rateStat
		for _, ip := range n.networkSpec.GetIPsByDeviceName(s2.Name) {
			rateStats[ip] = rateStat
		}
	}

	n.latestNetworkRateStat = rateStats

	return nil
}

func (n *NetworkRateProvider) GetNetworkRateStats() (NetworkRateStats) {
	return n.latestNetworkRateStat
}


