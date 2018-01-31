package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/spidermesh/spidermesh/cache"
	"github.com/spidermesh/spidermesh/config"
	"github.com/spidermesh/spidermesh/config/parse"
	"github.com/spidermesh/spidermesh/proxy"
	svcd "github.com/spidermesh/spidermesh/sd"

	"github.com/urfave/negroni"
)

func ListenAndServe(c *cache.TopCache, listen config.Listen) {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		var ip string
		var port int
		var svc *svcd.ServiceInfo
		servicePort := req.Header.Get("ServicePort")
		if len(servicePort) != 0 {
			req.Header.Del("ServicePort")
			ip = "localhost"
			port, _ = strconv.Atoi(servicePort)
		} else {
			service := strings.Split(req.Host, ".")[0]
			svc = c.GetService(service)
			if svc == nil {
				fmt.Fprintf(w, fmt.Sprintf("Couldn't find Service %s!", service))
				return
			} else {
				ip = svc.IP
				if listen.Route2Incoming {
					port = listen.Port
				} else {
					port = svc.Port
				}
			}
		}
		req.URL, _ = url.ParseRequestURI(
			fmt.Sprintf("%s://%s:%d%s", listen.Scheme, ip, port, req.URL.Path))
		if svc != nil {
			req.Header.Set("ServicePort", fmt.Sprintf("%d", svc.Port))
		}
		proxy.Forward(w, req)
	})

	n := negroni.New()
	recovery := negroni.NewRecovery()
	recovery.PrintStack = false
	n.Use(recovery)
	n.Use(negroni.NewLogger())
	n.UseHandler(mux)
	n.Run(fmt.Sprintf(":%d", listen.Port))
}

func main() {
	cfg := parse.Parse("./config/mesh.yml")
	sd := svcd.NewSD(&cfg.SD)
	listens := cfg.Listens
	config.CServices = config.RWServices{Services: cfg.Services}
	c := cache.NewCache(sd)
	c.Setup()
	var wg sync.WaitGroup
	for _, listen := range listens {
		wg.Add(1)
		go func(c *cache.TopCache, listen config.Listen) {
			defer wg.Done()
			ListenAndServe(c, listen)
		}(c, listen)
	}
	wg.Wait()
	log.Println("Done")
}
