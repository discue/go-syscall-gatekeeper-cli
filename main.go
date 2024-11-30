// Package main is the entry point for the k6 CLI application. It assembles all the crucial components for the running.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/discue/go-syscall-gatekeeper/app/uroot"
	"github.com/discue/go-syscall-gatekeeper/app/utils"
)

var logger = utils.NewLogger("main")

func main() {
	mainCtx, cancel := context.WithCancel(context.Background())

	tracee := startTracee(mainCtx)
	waitForShutdown(cancel, tracee)
}

func startTracee(c context.Context) context.Context {
	traceeCtx, _ := context.WithCancel(c)
	flag.Parse()
	args := flag.Args()

	cmd, err := uroot.Exec(traceeCtx, args[0], args[1:])
	if err != nil {
		fmt.Println(err.Error())
	}

	ctx, cancel := context.WithCancel(c)
	go func() {

		for {
			time.Sleep(1 * time.Second)
			var status syscall.WaitStatus
			p, _ := syscall.Wait4(cmd.Process.Pid, &status, syscall.WNOHANG, nil)
			if p != 0 {
				cancel()
				break
			}
		}

	}()

	return ctx
}

func waitForShutdown(cancel context.CancelFunc, tracee context.Context) {
	signal, stop := signal.NotifyContext(context.Background(), syscall.SIGKILL, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-signal.Done():
		logger.Info("Signal received. Stopping.")
		break
	case <-tracee.Done():
		logger.Info("Tracee stopped. Stopping.")
		break
	}

	cancel()
	time.Sleep(1 * time.Second)

	stop()

	logger.Info("Signal received. Stopping tracee.")
	logger.Info("Signal received. Waiting for tracee to stop.")
	<-tracee.Done()
	logger.Info("Shutting down.")

	os.Exit(0)
}
