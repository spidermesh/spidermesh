package cache

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/spidermesh/spidermesh/config"
	"github.com/spidermesh/spidermesh/sd"
)

type SubCache struct {
	sync.RWMutex
	Sign     string
	SvcInfos []*sd.ServiceInfo
}

type TopCache struct {
	sync.RWMutex
	Svcd    sd.ServiceDiscovery
	TagSvcs map[string]*SubCache
}

func NewCache(svcd sd.ServiceDiscovery) *TopCache {
	return &TopCache{
		Svcd:    svcd,
		TagSvcs: map[string]*SubCache{},
	}
}

func (t *TopCache) Setup() error {
	go func(t *TopCache) {
		for _ = range time.Tick(time.Second * 2) {
			t.UpdateSvcs()
		}
	}(t)
	return nil
}

func (t *TopCache) updateSvcInfos(s *SubCache, svcInfos []*sd.ServiceInfo, svcSign string) error {
	s.Lock()
	defer s.Unlock()
	s.SvcInfos = svcInfos
	s.Sign = svcSign
	return nil
}

func (t *TopCache) UpdateSvcsForTag(svcTag string) []*sd.ServiceInfo {
	st := strings.Split(svcTag, "-")
	svcInfos := t.Svcd.GetServicesByTag(st[0], st[1])
	if svcInfos != nil && len(svcInfos) != 0 {
		infos := []string{}
		for _, info := range svcInfos {
			infos = append(infos, fmt.Sprintf("%s", info))
		}
		sort.Strings(infos)
		svcSign := strings.Join(infos, "_")
		t.RLock()
		s, ok := t.TagSvcs[svcTag]
		t.RUnlock()
		if ok {
			s.RLock()
			if s.Sign == svcSign {
				// no update of micro services
				defer s.RUnlock()
			} else {
				s.RUnlock()
				t.updateSvcInfos(s, svcInfos, svcSign)
				log.Printf("Update svcInfos for %s with sign %s", svcTag, svcSign)
			}
		} else {
			t.Lock()
			defer t.Unlock()
			t.TagSvcs[svcTag] = &SubCache{
				Sign:     svcSign,
				SvcInfos: svcInfos,
			}
			log.Printf("Add %s with sign %s", svcTag, svcSign)
		}
		return svcInfos
	} else {
		return nil
	}
}

// for periodic update service info
func (t *TopCache) UpdateSvcs() error {
	for svcTag, _ := range t.TagSvcs {
		t.UpdateSvcsForTag(svcTag)
	}
	return nil
}

func (t *TopCache) GetServicesByTag(svcTag string) []*sd.ServiceInfo {
	t.RLock()
	if s, ok := t.TagSvcs[svcTag]; ok {
		t.RUnlock()
		s.RLock()
		defer s.RUnlock()
		return s.SvcInfos
	} else {
		t.RUnlock()
		return t.UpdateSvcsForTag(svcTag)
	}
}

func (t *TopCache) GetService(svc string) *sd.ServiceInfo {
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
	svcInfos := t.GetServicesByTag(fmt.Sprintf("%s-%s", svc, tag))
	if svcInfos == nil || len(svcInfos) == 0 {
		return nil
	}
	return svcInfos[rand.Intn(len(svcInfos))]
}
