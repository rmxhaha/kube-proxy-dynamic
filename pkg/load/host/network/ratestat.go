package network

import "time"

type NetworkRateStat struct {
	Name        string `json:"name"`        // interface name
	BytesSentRate   uint64 `json:"bytesSentRate"`   // number of bytes sent
	BytesRecvRate   uint64 `json:"bytesRecvRate"`   // number of bytes received
	PacketsSentRate uint64 `json:"packetsSentRate"` // number of packets sent
	PacketsRecvRate uint64 `json:"packetsRecvRate"` // number of packets received
	BytesSpeed      uint64 `json:"byteSpeed"`       // interface max speed
	NaiveUsage		float64 `json:"naiveUsage"`
	DateRecorded time.Time `json:"dateRecorded"` // when
}

type NetworkRateStats map[string]*NetworkRateStat

