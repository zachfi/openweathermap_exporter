FROM alpine:3.16
COPY openweathermap_exporter /bin/openweathermap_exporter
RUN chmod 0700 /bin/openweathermap_exporter
RUN apk --update add ca-certificates tzdata
ENTRYPOINT ["/bin/openweathermap_exporter"]
