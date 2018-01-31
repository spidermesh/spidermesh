package sd

import (
	"fmt"

	"github.com/spidermesh/spidermesh/config"
)

type ServiceInfo struct {
	IP   string
	Port int
}

func (s *ServiceInfo) String() string {
	return fmt.Sprintf("%s-%d", s.IP, s.Port)
}

type ServiceDiscovery interface {
	GetServicesByTag(string, string) []*ServiceInfo
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
