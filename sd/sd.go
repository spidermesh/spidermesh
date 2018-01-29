package sd

import (
	"spidermesh/config"
)

type ServiceInfo struct {
	IP   string
	Port int
}

type ServiceDiscovery interface {
	GetServices(string) []*ServiceInfo
}

type Factory func(sdc *config.SD) ServiceDiscovery

var (
	sdAgents = map[string]Factory{
		"consul": NewConsulAgent,
	}
)

func NewSD(sdc *config.SD) ServiceDiscovery {
	factory, ok := sdAgents[sdc.Agent]
	if !ok {
		return nil
	}
	return factory(sdc)
}
