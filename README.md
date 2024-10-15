# OpenTelemetry Example

OpenTelemetry good to send send trace and metric data to different platform.

| Name           | URL                                   |
|----------------|---------------------------------------|
| grafana        | http://localhost:3000                 |
| prometheus     | http://localhost:9090                 |
| otel-collector | http://localhost:8889/metrics         |
| tempo          | http://localhost:3200                 |
| example        | http://localhost:8080/api/v1/swagger/ |

## Quick Start

Initialize compose-file

```sh
make env
```

After that run this example program, run command set the otel env variables to open telemetry connection in code.

```sh
make run
```

<details><summary>Details</summary>

In prometheus go to status -> targets to check tartgets health.

</details>

## Metric / Trace

Check the https://github.com/worldline-go/tell
