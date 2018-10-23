package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/oneiro-ndev/ndau/pkg/ndauapi/cfg"
	"github.com/oneiro-ndev/ndau/pkg/ndauapi/svc"

	log "github.com/sirupsen/logrus"
)

func usage() {
	fmt.Fprintf(os.Stderr, `
ndauapi is a simple http server for interacting with nodes.

Usage

Environment variables

	Log level. (default: info)
	[NDAUAPI_LOG_LEVEL=(error|warn|info|debug)]

	Port where this API will listen. (default: 3030)
	[NDAUAPI_PORT=<3030>]

	Node address.
	NDAUAPI_NODE_ADDRESS=<http://your_node_ip:your_rpc_port>

Flags

	-docs Generates boneful API documentation in markdown.

Example

	NDAUAPI_NODE_ADDRESS=http://127.0.0.1:26658 ./ndauapi [-docs]

`)
}

type siglistener struct {
	sigchan chan os.Signal
}

func (s *siglistener) watchSignals() {
	go func() {
		if s.sigchan == nil {
			s.sigchan = make(chan os.Signal, 1)
		}
		signal.Notify(s.sigchan, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)
		for {
			sig := <-s.sigchan
			switch sig {
			// case syscall.SIGHUP:
			// s.Logger.Println("Got refresh request (SIGHUP) -- Refreshing vasco.")
			// s.Refresh()
			case syscall.SIGTERM, syscall.SIGINT:
				// s.Logger.Println("Unregistering before shutting down.")
				// s.Unregister()
				os.Exit(0)
			}
		}
	}()
}

func main() {

	// handle flags to generate docs
	docsFlag := flag.Bool("docs", false, "Prints API documents to stdout.")
	flag.Parse()
	if *docsFlag {
		svc := svc.New(cfg.Cfg{})
		svc.GenerateDocumentation(os.Stdout)
		return
	}

	// initialize configuration
	cf, warn, err := cfg.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not get config: %v\n", err)
		usage()
		os.Exit(1)
	}
	if len(warn) != 0 {
		fmt.Fprintf(os.Stderr, "config warning: %v\n", strings.Join(warn, ", "))
	}

	fmt.Fprintf(os.Stderr, "      █                  █\n █   ██  █  █ █  █   █\n█ █ █ █ █ █ █ █ █ █ █ █  █\n█ █  ██  ██  ██  ██ ██   █\n                    █\n")
	log.Printf("server listening on port %v\n", cf.Port)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", cf.Port),
		Handler: svc.NewLogMux(cf),
	}
	sl := &siglistener{}
	sl.watchSignals()
	log.Fatal(server.ListenAndServe())
}
