package ipvs

import (
	"github.com/docker/libnetwork/ipvs"
	"fmt"
)

type Destinations map[string]*ipvs.Destination

type VirtualServer struct {
	IPVSService *ipvs.Service
	IPVSDestinations Destinations
}

func NewVirtualServer(IPVSService *ipvs.Service) *VirtualServer {
	return &VirtualServer{IPVSService: IPVSService, IPVSDestinations: Destinations{}}
}

func (vs *VirtualServer) AddDestination(d *ipvs.Destination){
	vs.IPVSDestinations[getDestinationID(d)] = d
}

type VirtualServers map[string]*VirtualServer

func (vss *VirtualServers) AddVirtualServer(server *VirtualServer){
	(*vss)[getServiceID(server.IPVSService)] = server
}



func getServiceID(s *ipvs.Service) string {
	return fmt.Sprintf("%s%d%d", s.Address.String(), s.Protocol, s.Port)
}

func getDestinationID(s *ipvs.Destination) string {
	return fmt.Sprintf("%d:%d", s.Address, s.Port)
}
