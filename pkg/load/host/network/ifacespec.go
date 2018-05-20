package network

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

type InterfaceSpec struct {
	InterfaceName string
	Speed *resource.Quantity
	IPAddresses []string
}

