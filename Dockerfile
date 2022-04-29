FROM alpine:3.15

COPY docker_service_exporter /bin/docker_service_exporter

RUN chmod 755 /bin/docker_service_exporter

EXPOSE      9115
USER        nobody
ENTRYPOINT  [ "/bin/docker_service_exporter" ]
