package sd

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"spidermesh/config"

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

func (c *Consul) getServicesByTag(svc, tag string) []*ServiceInfo {
	services, _, err := c.client.Catalog().Service(svc, tag, nil)
	if len(services) == 0 || err != nil {
		log.Printf("Can't find service %s\n", svc)
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

func (c *Consul) GetServices(svc string) []*ServiceInfo {
	tag := ""
	services := config.CServices.GetServices(svc)
	if len(services) > 0 {
		ratios := []int{}
		tags := []string{}
		sum := 0
		for _, service := range services {
			sum += service.Weight
			ratios = append(ratios, sum)
			tags = append(tags, service.Tag)
		}
		rand.Seed(time.Now().Unix())
		randInt := rand.Intn(sum)
		for i, ratio := range ratios {
			if randInt < ratio {
				tag = tags[i]
				break
			}
		}
	}
	svcInfos := c.getServicesByTag(svc, tag)
	if svcInfos == nil || len(svcInfos) == 0 {
		return nil
	}
	//	return svcInfos[rand.Intn(len(svcInfos))]
}
