package config

import (
	"time"

	"gitlab.test.igdcs.com/finops/nextgen/utils/metrics/tell"
)

var (
	AppName    = "telemetry"
	AppVersion = "v0.0.0"
	LoadName   = ""
	StartDate  = time.Now()
)

type Prefix struct {
	Vault  string `cfg:"vault"`
	Consul string `cfg:"consul"`
}

var LoadConfig = struct {
	Prefix  Prefix `cfg:"prefix"`
	AppName string `cfg:"app_name"`
}{
	AppName: AppName,
}

var Application = struct {
	LogLevel  string `cfg:"log_level" default:"info"`
	Host      string `cfg:"host" default:"0.0.0.0"`
	Port      string `cfg:"port" default:"8080"`
	BasePath  string `cfg:"base_path"`
	Telemetry tell.Config
}{}
