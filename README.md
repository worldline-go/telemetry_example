# OpenTelemetry Example

OpenTelemetry good to send send trace and metric data to different platform.

Some notes how we do that:

__-__ For tracing we can use jaeger and elastic's APM, both of support grpc and http to receive OTPL (open telemetry protocol).  
__-__ We can directly convert data for jaeger and send directly jaeger's port but looks old method don't do that.  
__-__ I guess we don't need to investigate all of request and we can put some query or variable in grpc, kafka and tell services trace this specific request.  
__-__ Opentelemetry has collector and we can send all of our datas in there after that we can configure to redirect metrics to prometheus and redirect trace to other tools also it is enable us to work on different tools.  
__-__ Don't close program if telemetry cannot connect to collector, skip and try in background.  
__-__ OTEL_EXPORTER_OTLP_ENDPOINT use for all services to show collector's GRPC endpoint otel-collector:4317  


| Name       | URL                                   |
|------------|---------------------------------------|
| grafana    | http://localhost:3000                 |
| jaeger     | http://localhost:16686                |
| zipkin     | http://localhost:9411                 |
| cadvisor   | http://localhost:9091                 |
| prometheus | http://localhost:9090                 |
| example    | http://localhost:8080/api/v1/swagger/ |

Initialize compose-file

```sh
make env
```

After that run this example program

```sh
make run
```

Go to localhost 3000 for grafana and login with `admin:admin`.

Add first datasource to show our promethues URL (9090).

Click dashboard and show custom metrics in there.

For testing import cadvisor's dashboard 14282 and select prometheus.

In prometheus go to status -> targets to check tartgets health.

## Test ENV's prometheus settings

If you want to add configuration in our test environment:

```sh
ssh am2vm2300.test.igdcs.com
```

Add configuration in here

```
/export/config/prometheus/targets
```

Still need to someone reload/restart prometheus to detect new configuration.

## Resources

https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/examples/demo  
