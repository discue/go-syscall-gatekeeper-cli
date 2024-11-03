package main

type GatekeeperServerConfig struct {
	ServerPort int `split_words:"true" default:"3333"`
}

type LivenessProxyConfig struct {
	LivenessTargetProtocol string `split_words:"true" default:"http"`
	LivenessTargetHostname string `split_words:"true" default:"127.0.0.1"`
	LivenessTargetPort     int    `split_words:"true" default:"8080"`
	LivenessTargetPath     string `split_words:"true" default:"/live"`
}

type HealthProxyConfig struct {
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
