package config

import "time"

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
	LogLevel  string `cfg:"log_level"`
	Host      string `cfg:"host"`
	Port      string `cfg:"port"`
	BasePath  string `cfg:"base_path"`
	Collector string `cfg:"collector" env:"OTEL_EXPORTER_OTLP_ENDPOINT"`
}{
	LogLevel:  "info",
	Host:      "0.0.0.0",
	Port:      "8080",
	Collector: "localhost:4317",
}
