// Package main is the entry point for the k6 CLI application. It assembles all the crucial components for the running.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	server "github.com/discue/go-syscall-gatekeeper/app/proxy"
	"github.com/discue/go-syscall-gatekeeper/app/uroot"
	"github.com/stfsy/go-api-kit/utils"
)

var logger = utils.NewLogger("main")

func main() {
	c := context.Background()
	startTracee(c)
	startServer()
	stopServerAfterSignal(c)
}

func startTracee(c context.Context) {
	flag.Parse()
	args := flag.Args()

	go func() {
		err := uroot.Exec(c, args[0], args[1:])
		if err != nil {
			fmt.Println(err.Error())
		}
	}()
}

func startServer() {
	go func() {
		server.Start()
	}()
}

func stopServerAfterSignal(c context.Context) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	c.Done()
	server.Stop()

	logger.Info("Graceful shutdown complete.")
}
