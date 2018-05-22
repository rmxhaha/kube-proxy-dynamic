package disk

import (
	utildisk "github.com/shirou/gopsutil/disk"
	"time"
)

type DiskUtilStat map[string]float64

type DiskStatProvider struct {
	lastDiskStats map[string]uint64
	lastCollect time.Time
	updateInterval time.Duration
	latestDiskUtilStat DiskUtilStat
}

func (p *DiskStatProvider) populate() error {
	m, err := utildisk.IOCounters("sda","vda")
	if err != nil {
		return err
	}

	ret := make(map[string]uint64, 0)

	for name, stat := range m {
		ret[name] = stat.IoTime
	}

	p.lastDiskStats = ret
	p.lastCollect = time.Now()
	return nil
}

func (p *DiskStatProvider) update() error {
	lastStats := p.lastDiskStats
	lastCollect := p.lastCollect

	if err := p.populate(); err != nil {
		return err
	}

	currentStats := p.lastDiskStats
	currentTime := p.lastCollect

	observeDuration := currentTime.Sub(lastCollect)

	diskstat := DiskUtilStat{}

	for name, cstat := range currentStats {
		if prevstat, ok := lastStats[name]; ok {
			utilization := float64(cstat - prevstat) / observeDuration.Seconds()
			diskstat[name] = utilization
		}
	}

	p.latestDiskUtilStat = diskstat
	return nil
}

func (p *DiskStatProvider) GetDiskUtilStat() DiskUtilStat {
	return p.latestDiskUtilStat
}

func (p *DiskStatProvider) updateLoop(){
	for {
		p.update()
		time.Sleep(p.updateInterval)
	}
}

func NewDiskStatProvider(updateInterval time.Duration) (*DiskStatProvider, error) {
	p := &DiskStatProvider{ updateInterval: updateInterval }
	if err := p.populate(); err != nil {
		return nil, err
	}

	go p.updateLoop()

	return p, nil
}