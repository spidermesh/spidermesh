package sd

import (
	"fmt"
	"log"

	"github.com/spidermesh/spidermesh/config"

	consul "github.com/hashicorp/consul/api"
)

type Consul struct {
	IP         string
	Port       int
	Datacenter string
	client     *consul.Client
}

func NewConsulAgent(sdc *config.SD) ServiceDiscovery {
	agent := &Consul{
		IP:         sdc.Addr,
		Port:       sdc.Port,
		Datacenter: sdc.Datacenter,
	}
	err := agent.PrepareClient()
	if err != nil {
		return nil
	}
	return agent
}

func (c *Consul) PrepareClient() error {
	if c.client == nil {
		cfg := consul.DefaultConfig()
		cfg.Address = fmt.Sprintf("%s:%d", c.IP, c.Port)
		cfg.Datacenter = c.Datacenter

		client, err := consul.NewClient(cfg)
		if err != nil {
			return err
		}
		c.client = client
	}
	return nil
}

func (c *Consul) GetServicesByTag(svc, tag string) []*ServiceInfo {
	services, _, err := c.client.Catalog().Service(svc, tag, nil)
	if len(services) == 0 || err != nil {
		// log.Printf("Can't find service %s\n", svc)
		if err != nil {
			log.Println(err)
		}
		return nil
	}
	svcInfos := []*ServiceInfo{}
	for _, service := range services {
		svcInfos = append(svcInfos, &ServiceInfo{
			IP:   service.Address,
			Port: service.ServicePort,
		})
	}
	return svcInfos
}
