FROM alpine:3.15 as certs
COPY ./bin/linux/openweathermap_exporter /bin/openweathermap_exporter
RUN chmod 0700 /bin/openweathermap_exporter
RUN mkdir /var/openweathermap_exporter
RUN apk --update add ca-certificates
RUN apk add libc6-compat
RUN apk add tzdata
ENTRYPOINT ["/bin/openweathermap_exporter"]
