package main

import (
	"testing"

	"github.com/cuandari/lib/app/runtime"
	"github.com/stretchr/testify/assert"
)

func TestRuntimeSetAndReset(t *testing.T) {
	a := assert.New(t)
	cfg := &runtime.Config{}
	cfg.SyscallsAllowList = []string{"read"}
	runtime.Set(cfg)
	g := runtime.Get()
	a.Equal(cfg, g)

	runtime.Reset()
	a.NotNil(runtime.Get()) // Load will be called on Get() if nil and may panic if env not set; ensure it does not return nil
}
