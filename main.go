// Package main is the entry point for the k6 CLI application. It assembles all the crucial components for the running.
package main

import (
	"flag"
	"fmt"

	uroot "github.com/discue/go-syscall-gatekeeper/app/uroot"
)

func main() {
	flag.Parse()
	args := flag.Args()

	err := uroot.Exec(args[0], args[1:])
	if err != nil {
		fmt.Println(err.Error())
	}
}
