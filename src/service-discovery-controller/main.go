package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"service-discovery-controller/addresstable"
	"service-discovery-controller/config"
	"syscall"
)

type host struct {
	IPAddress       string                 `json:"ip_address"`
	LastCheckIn     string                 `json:"last_check_in"`
	Port            int32                  `json:"port"`
	Revision        string                 `json:"revision"`
	Service         string                 `json:"service"`
	ServiceRepoName string                 `json:"service_repo_name"`
	Tags            map[string]interface{} `json:"tags"`
}

type registration struct {
	Hosts   []host `json:"hosts"`
	Env     string `json:"env"`
	Service string `json:"service"`
}

func main() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGTERM, os.Interrupt)
	configPath := flag.String("c", "", "path to config file")
	flag.Parse()

	bytes, err := ioutil.ReadFile(*configPath)
	if err != nil {
		fmt.Printf("Could not read config file at path '%s'", *configPath)
		os.Exit(2)
	}

	config, err := config.NewConfig(bytes)
	if err != nil {
		fmt.Printf("Could not parse config file at path '%s'", *configPath)
		os.Exit(2)
	}

	address := fmt.Sprintf("%s:%s", config.Address, config.Port)
	l, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Address (%s) not available", address))
		os.Exit(1)
	}

	addressTable := addresstable.NewAddressTable()

	addressTable.Add([]string{"app-id.internal.local."}, "192.168.0.1")
	addressTable.Add([]string{"app-id.internal.local."}, "192.168.0.2")

	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.1")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.2")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.3")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.4")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.5")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.6")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.7")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.8")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.9")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.10")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.11")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.12")
	addressTable.Add([]string{"large-id.internal.local."}, "192.168.0.13")

	go func() {
		http.Serve(l, http.HandlerFunc(func(resp http.ResponseWriter, req *http.Request) {
			serviceKey := path.Base(req.URL.Path)

			ips := addressTable.Lookup(serviceKey)
			hosts := []host{}
			for _, ip := range ips {
				hosts = append(hosts, host{
					IPAddress: ip,
					Tags:      make(map[string]interface{}),
				})
			}

			var err error
			json, err := json.Marshal(registration{Hosts: hosts})
			if err != nil {
				http.Error(resp, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = resp.Write(json)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error writing to http response body") // not tested
			}
		}))
	}()

	fmt.Println("Server Started")

	select {
	case <-signalChannel:
		fmt.Println("Shutting service-discovery-controller down")
		return
	}
}
