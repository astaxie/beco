package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"
)

// VERSION holds the beco version
const (
	VERSION = "0.0.1"
)

func main() {
	var (
		config = flag.String("config", filepath.FromSlash("/etc/beco.toml"),
			"config for the gateway server. If it's not setted, it will load the /etc/beco.toml")
		cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
		memprofile = flag.String("memprofile", "", "write memory profile to this file")
		version    = flag.Bool("version", false, "Output version and exit")
	)
	if *version {
		fmt.Println(VERSION)
		os.Exit(0)
	}
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()
	if *config == "" {
		flag.Usage()
		return
	}
	conf, err := parseConifg(*config)
	if err != nil {
		log.Fatal("parseConfig:", err)
	}
	if *memprofile != "" {
		profileMEM(*memprofile)
	}

	if *cpuprofile != "" {
		profileCPU(*cpuprofile)
	}
	server, err := NewServer()
	if err != nil {
		log.Fatal("NewServer:", err)
	}
	for _, proxy := range conf.Proxys {
		hnd, err := ProxyHandler(proxy)
		if err != nil {
			log.Fatal("Create backend Proxy:", err)
		}
		server.Handle(proxy.Prefix, hnd)
	}
	ending := make(chan error, 1)
	go func() {
		if err = server.Run(fmt.Sprintf("%s:%d", conf.Addr, conf.Port)); err != nil {
			log.Fatal("Server running:", err)
		}
		ending <- err
	}()
	if conf.SSL.Cert != "" && conf.SSL.Key != "" && conf.SSL.Port != 0 {
		go func() {
			if err = server.RunTLS(fmt.Sprintf("%s:%d", conf.SSL.Addr, conf.SSL.Port), conf.SSL.Cert, conf.SSL.Key); err != nil {
				log.Fatal("Server running:", err)
			}
			ending <- err
		}()
	}
	err = <-ending
	log.Fatal("Run Server err:", err)
}

func profileCPU(cpuprofile string) {
	if cpuprofile != "" {
		f, err := os.Create(cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)

		time.AfterFunc(30*time.Second, func() {
			pprof.StopCPUProfile()
			f.Close()
			log.Println("Stop profiling after 30 seconds")
		})
	}
}

func profileMEM(memprofile string) {
	if memprofile != "" {
		f, err := os.Create(memprofile)
		if err != nil {
			log.Fatal(err)
		}
		time.AfterFunc(30*time.Second, func() {
			pprof.WriteHeapProfile(f)
			f.Close()
		})
	}
}
