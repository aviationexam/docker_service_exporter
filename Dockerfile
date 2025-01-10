FROM golang:1.23.4-alpine3.19 as builder

WORKDIR /src

COPY src .

ARG APP_VERSION
ARG APP_BUILD_DATE
ARG APP_REVISION
ARG APP_BRANCH

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="\
    -w\
    -s\
    -X 'main.Version=${APP_VERSION}'\
    -X 'app/build.Time=$( date '+%F %H-%M-%S' )'\
    -X 'github.com/prometheus/common/version.Version=${APP_VERSION}'\
    -X 'github.com/prometheus/common/version.Revision=${APP_REVISION}'\
    -X 'github.com/prometheus/common/version.Branch=${APP_BRANCH}'\
    -X 'github.com/prometheus/common/version.BuildDate=$( date '+%F %H-%M-%S' )'\
    " .

FROM alpine:3.21.2

COPY --from=builder /src/docker_service_exporter /bin/docker_service_exporter

EXPOSE      9115

# required to access the /var/run/docker.sock
USER        root

ENTRYPOINT  [ "/bin/docker_service_exporter" ]
