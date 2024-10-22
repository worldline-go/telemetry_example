# OpenTelemetry Example

OpenTelemetry good to send send trace and metric data to different platform.

| Name                 | URL                                   |
| -------------------- | ------------------------------------- |
| grafana              | http://localhost:3000                 |
| prometheus           | http://localhost:9090                 |
| otel-collector       | http://localhost:8889/metrics         |
| otel-collector graph | http://localhost:8890/metrics         |
| tempo                | http://localhost:3200                 |
| example              | http://localhost:8080/api/v1/swagger/ |

![services](./_assets/services.excalidraw.svg)

## Quick Start

Initialize compose-file for development environment.

```sh
make env
```

Start the services.

```sh
make env-services
```

Stop the services.

```sh
make env-services-destroy
make env-destroy
```

## Metric / Trace

Check the https://github.com/worldline-go/tell
