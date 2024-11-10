package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type GatekeeperServerConfig struct {
	GatekeeperServerEnabled bool `split_words:"true" default:"true"`
	GatekeeperServerPort    int  `split_words:"true" default:"8081"`
}

type LivenessProxyConfig struct {
	LivenessTargetEnabled  bool   `split_words:"true" default: "false"`
	LivenessTargetProtocol string `split_words:"true" default:"http"`
	LivenessTargetHostname string `split_words:"true" default:"127.0.0.1"`
	LivenessTargetPort     int    `split_words:"true" default:"8080"`
	LivenessTargetPath     string `split_words:"true" default:"/live"`
}

type HealthProxyConfig struct {
	HealthTargetEnabled  bool   `split_words:"true" default: "false"`
	HealthTargetProtocol string `split_words:"true" default:"http"`
	HealthTargetHostname string `split_words:"true" default:"127.0.0.1"`
	HealthTargetPort     int    `split_words:"true" default:"8080"`
	HealthTargetPath     string `split_words:"true" default:"/health"`
}

type Config struct {
	HealthProxyConfig
	LivenessProxyConfig
	GatekeeperServerConfig
}

var c *Config

func Load() {
	var s Config
	err := envconfig.Process("GATEKEEPER", &s)
	if err != nil {
		panic(fmt.Sprintf("unable to read environment configuration %s", err.Error()))
	}

	c = &s
}

func Get() Config {
	if c == nil {
		Load()
	}

	return *c
}

func reset() {
	c = nil
}
