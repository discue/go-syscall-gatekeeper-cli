// Package main is the entry point for the k6 CLI application. It assembles all the crucial components for the running.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	server "github.com/discue/go-syscall-gatekeeper/app/proxy"
	"github.com/discue/go-syscall-gatekeeper/app/uroot"
	"github.com/discue/go-syscall-gatekeeper/app/utils"
)

var logger = utils.NewLogger("main")

func main() {
	mainCtx, cancel := context.WithCancel(context.Background())

	tracee := startTracee(mainCtx)
	startServer()
	stopServerAfterSignal(cancel, tracee)
}

func startTracee(c context.Context) *exec.Cmd {
	traceeCtx, _ := context.WithCancel(c)
	flag.Parse()
	args := flag.Args()

	cmd, err := uroot.Exec(traceeCtx, args[0], args[1:])
	if err != nil {
		fmt.Println(err.Error())
	}

	return cmd
}

func startServer() {
	go func() {
		server.Start()
	}()
}

func stopServerAfterSignal(cancel context.CancelFunc, tracee *exec.Cmd) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()

	logger.Info("Signal received. Stopping.")

	cancel()
	time.Sleep(1 * time.Second)

	stop()
	server.Stop()

	logger.Info("Signal received. Stopping tracee.")
	logger.Info("Signal received. Waiting for tracee to stop.")
	tracee.Wait()
	logger.Info("Shutting down.")

	os.Exit(0)
}
