FROM scratch

COPY telemetry /telemetry

ENTRYPOINT [ "/telemetry" ]
