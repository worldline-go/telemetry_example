# OpenTelemetry Example

OpenTelemetry good to send send trace and metric data to different platform.

| Name       | URL                                   |
|------------|---------------------------------------|
| grafana    | http://localhost:3000                 |
| jaeger     | http://localhost:16686                |
| zipkin     | http://localhost:9411                 |
| cadvisor   | http://localhost:9091                 |
| prometheus | http://localhost:9090                 |
| example    | http://localhost:8080/api/v1/swagger/ |

<details><summary>Initialize</summary>

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

</details>

<details><summary>Test ENV's prometheus settings</summary>

If you want to add configuration in our test environment:

```sh
ssh am2vm2300.test.igdcs.com
```

Add configuration in here

```
/export/config/prometheus/targets
```

Still need to someone reload/restart prometheus to detect new configuration.

</details>

## Metric

Some metrics directly comes from environment variable special for opentelemetry collector but this is experimental and just works with collector, best use case declare own standart.

Best set service name service version and service instanance-id in own library.

In the scaling services metrics should have `service.name`, and uniq `service.instance.id`. Opentelemetry detect these kind of arguments and well parse and set as resource for you.  
We can use swarm's own go templates for this kind of job.

Example in run:

```sh
OTEL_RESOURCE_ATTRIBUTES=service.name=telemetry,service.instance.id=xyz123,service.namespace=transaction,service.version=v1.0.0 make run
```

Check much more details of attributes in here https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md

In metrics you give more than one reader in provider, same time can send to collector and show in `/metrics`.

## Trace



## Notes

Some notes how we do that:

__-__ For tracing we can use jaeger and elastic's APM, both of support grpc and http to receive OTPL (open telemetry protocol).  
__-__ We can directly convert data for jaeger and send directly jaeger's port but looks old method don't do that.  
__-__ I guess we don't need to investigate all of request and we can put some query or variable in grpc, kafka and tell services trace this specific request.  
__-__ Opentelemetry has collector and we can send all of our datas in there after that we can configure to redirect metrics to prometheus and redirect trace to other tools also it is enable us to work on different tools.  
__-__ Don't close program if telemetry cannot connect to collector, skip and try in background.  
__-__ OTEL_EXPORTER_OTLP_ENDPOINT use for all services to show collector's GRPC endpoint otel-collector:4317  

## Resources

https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/examples/demo  
https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/resource/semantic_conventions/README.md  
https://docs.docker.com/engine/swarm/services/#create-services-using-templates
