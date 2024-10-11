FROM scratch

COPY /dist/telemetry /telemetry

ENTRYPOINT [ "/telemetry" ]
