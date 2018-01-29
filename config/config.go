package config

import (
	"sync"
)

var (
	CServices RWServices
)

type SD struct {
	Agent      string `yaml:"agent"`
	Addr       string `yaml:"addr"`
	Port       int    `yaml:"port"`
	Datacenter string `yaml:"datacenter"`
}

type Listen struct {
	Port           int    `yaml:"port"`
	Scheme         string `yaml:"scheme"`
	Route2Incoming bool   `yaml:"route2incoming"`
}

type Service struct {
	Name   string `yaml:"name"`
	Tag    string `yaml:"tag"`
	Weight int    `yaml:"weight"`
}

type Services []Service

type Config struct {
	SD       `yaml:"sd"`
	Listens  []Listen `yaml:"listens,flow"`
	Services Services `yaml:"services,flow"`
}

type RWServices struct {
	sync.RWMutex
	Services Services
}

func (s *RWServices) GetServices(svc string) []*Service {
	services := []*Service{}
	s.RLock()
	defer s.RUnlock()
	for _, service := range s.Services {
		if svc == service.Name {
			services = append(services, &service)
		}
	}
	return services
}
