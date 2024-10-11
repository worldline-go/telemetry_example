# OpenTelemetry Example

OpenTelemetry good to send send trace and metric data to different platform.

| Name           | URL                                   |
|----------------|---------------------------------------|
| grafana        | http://localhost:3000                 |
| jaeger         | http://localhost:16686                |
| zipkin         | http://localhost:9411                 |
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

Run without opentelemetry

```sh
make run-without
```

<details><summary>Details</summary>

Go to localhost 3000 for grafana and login with `admin:admin`.

Add first datasource to show our promethues URL (9090).

Click dashboard and show custom metrics in there.

For testing import cadvisor's dashboard 14282 and select prometheus.

In prometheus go to status -> targets to check tartgets health.

</details>

## Metric / Trace

Check the https://github.com/worldline-go/tell
