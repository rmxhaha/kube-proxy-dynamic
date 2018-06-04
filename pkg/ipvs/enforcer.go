package ipvs

import (
	"github.com/docker/libnetwork/ipvs"
	"reflect"
)

type IPVSRuleEnforcer struct {
	handle *ipvs.Handle
	enforcedeleteservice bool
}

func (re *IPVSRuleEnforcer) getCurrentVirtualServers() (VirtualServers, error) {
	vss := VirtualServers{}

	services, err := re.handle.GetServices()
	if err != nil { return nil, err }


	for _, srv := range services {
		destinationarr, err := re.handle.GetDestinations(srv)

		if err != nil {
			return nil, err
		}

		vs := &VirtualServer{IPVSService: srv, IPVSDestinations: Destinations{}}
		for _, d := range destinationarr {
			vs.AddDestination(d)
		}

		vss.AddVirtualServer(vs)
	}

	return vss, nil
}


func setDifference(vss1 VirtualServers, vss2 VirtualServers) VirtualServers {
	result := VirtualServers{}

	for vsid, vs := range vss1 {
		if _, ok := vss2[vsid]; !ok {
			result[vsid] = vs
		}
	}

	return result
}


func setDifference2(ds1 Destinations, ds2 Destinations) Destinations {
	result := Destinations{}

	for vsid, ds := range ds1 {
		if _, ok := ds2[vsid]; !ok {
			result[vsid] = ds
		}
	}

	return result
}

func intersect(vss1 VirtualServers, vss2 VirtualServers) VirtualServers {
	result := VirtualServers{}

	for vsid, vs := range vss1 {
		if _, ok := vss2[vsid]; ok {
			result[vsid] = vs
		}
	}

	return result
}


func intersect2(vss1 Destinations, vss2 Destinations) Destinations {
	result := Destinations{}

	for dsid, vs := range vss1 {
		if _, ok := vss2[dsid]; ok {
			result[dsid] = vs
		}
	}

	return result
}

func NewEnforcer(enforcedeleteservice bool) (*IPVSRuleEnforcer, error) {
	handle, err := ipvs.New("")

	if err != nil { return nil, err }

	return &IPVSRuleEnforcer{ handle: handle, enforcedeleteservice: enforcedeleteservice }, nil
}

func (re *IPVSRuleEnforcer) Enforce(vss VirtualServers) error {
	var err error
	currentVSS, err := re.getCurrentVirtualServers()
	if err != nil {return err}

	toBeAdded := setDifference(vss, currentVSS)
	toBeDeleted := setDifference(currentVSS, vss)
	toBeDeeperChecked := intersect(vss, currentVSS)

	for _, s := range toBeAdded {
		err := re.handle.NewService(s.IPVSService)
		if err != nil { return err }

		for _, d := range s.IPVSDestinations {
			err := re.handle.NewDestination(s.IPVSService, d)
			if err != nil { return err }
		}
	}

	if re.enforcedeleteservice {
		for _, s := range toBeDeleted {
			err := re.handle.DelService(s.IPVSService)
			if err != nil { return err }
		}
	}

	for vsk := range toBeDeeperChecked {
		if !reflect.DeepEqual(currentVSS[vsk],vss[vsk]) {
			re.handle.UpdateService(vss[vsk].IPVSService)
		}

		toBeDeletedDestinations := setDifference2(currentVSS[vsk].IPVSDestinations, vss[vsk].IPVSDestinations)
		toBeAddedDestinations := setDifference2(vss[vsk].IPVSDestinations, currentVSS[vsk].IPVSDestinations)
		toBeUpdated := intersect2(vss[vsk].IPVSDestinations, currentVSS[vsk].IPVSDestinations)

		for _, ds := range toBeDeletedDestinations {
			err := re.handle.DelDestination(vss[vsk].IPVSService, ds)
			if err != nil { return err }
			//log.Printf("ds %s %d deleted", ds.Address, ds.Port)
		}

		for _, ds := range toBeAddedDestinations {
			err := re.handle.NewDestination(vss[vsk].IPVSService, ds)
			if err != nil { return err }
			//log.Printf("ds %s %d added", ds.Address, ds.Port)
		}

		for _, ds := range toBeUpdated {
			if !reflect.DeepEqual(vss[vsk].IPVSService, ds) {
				err := re.handle.UpdateDestination(vss[vsk].IPVSService, ds)
				if err != nil { return err }
			}
		}
	}

	return nil
}
