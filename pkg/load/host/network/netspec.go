package network

import (
	"net"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"github.com/prometheus/procfs/sysfs"
)

type NetworkSpec struct {
	interfacesLimitByName map[string]*InterfaceSpec
	interfacesLimitByIP map[string]*InterfaceSpec
}


// id could be deviceName or ipv4 or ipv6
func (l *NetworkSpec) GetSpeed(id string) *resource.Quantity{
	if l.interfacesLimitByName[id] != nil {
		return l.interfacesLimitByName[id].Speed
	} else if l.interfacesLimitByIP[id] == nil {
		return l.interfacesLimitByIP[id].Speed
	} else {
		return nil
	}
}

func (l *NetworkSpec) GetIPsByDeviceName(deviceName string) []string {
	if interfaceLimit, ok := l.interfacesLimitByName[deviceName]; ok {
		return interfaceLimit.IPAddresses
	} else {
		return nil
	}
}

func NewNetworkSpec() (*NetworkSpec, error) {
	networkLimit := &NetworkSpec{
		interfacesLimitByName: map[string]*InterfaceSpec{},
		interfacesLimitByIP: map[string]*InterfaceSpec{},
	}

	return networkLimit, nil
}

func (l *NetworkSpec) addInterface(limit *InterfaceSpec){
	l.interfacesLimitByName[limit.InterfaceName] = limit
	for _,ip := range limit.IPAddresses {
		l.interfacesLimitByIP[ip] = limit
	}
}

func NewHostNetworkSpec() (*NetworkSpec, error){
	ns, err := NewNetworkSpec()

	if err != nil {
		return nil, err
	}

	err = populateHostNetworkSpec(ns)
	if err != nil {
		return nil, err
	}

	return ns, nil
}

func populateHostNetworkSpec(l *NetworkSpec) error {
	netClass, err := sysfs.NewNetClass()
	if err != nil {
		return err
	}

	ifaces, err := net.Interfaces()

	if err != nil {
		return err
	}

	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			return err
		}

		var ips []string

		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			fmt.Printf("ip %s\n", ip.String())
			ips = append(ips, ip.String())
		}

		// net class give speed in MBit/s
		// converts to bytes/s
		speedBytes := netClass[i.Name].Speed * 1000000 / 8

		interfaceLimit := &InterfaceSpec{
			InterfaceName: i.Name,
			Speed: resource.NewQuantity(speedBytes, resource.BinarySI),
			IPAddresses: ips,
		}

		l.addInterface(interfaceLimit)
	}

	return nil
}


