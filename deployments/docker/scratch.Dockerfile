FROM scratch

COPY /dist/telemetry /telemetry
COPY /migrations /migrations

ENTRYPOINT [ "/telemetry" ]
