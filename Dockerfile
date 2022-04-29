FROM golang:1.18.1-alpine3.15 as builder

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
    -X 'app/build.Time=${APP_BUILD_DATE}'\
    -X 'github.com/prometheus/common/version.Version=${APP_VERSION}'\
    -X 'github.com/prometheus/common/version.Revision=${APP_REVISION}'\
    -X 'github.com/prometheus/common/version.Branch=${APP_BRANCH}'\
    -X 'github.com/prometheus/common/version.BuildDate=${APP_BUILD_DATE}'\
    " .

FROM alpine:3.15

COPY --from=builder /src/docker_service_exporter /bin/docker_service_exporter

EXPOSE      9115
USER        nobody
ENTRYPOINT  [ "/bin/docker_service_exporter" ]
