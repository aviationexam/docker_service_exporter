# Docker Service Exporter

Docker Service exporter for various metrics about Docker Services, written in Go.

### Installation

#### Docker

```bash
docker pull aviationexam/docker_service_exporter:latest
docker run --rm -p 9115:9115 aviationexam/docker_service_exporter:latest
```

Example `docker-compose.yml`:

```yaml
elasticsearch_exporter:
  image: aviationexam/docker_service_exporter:latest
  restart: always
  ports:
    - "127.0.0.1:9115:9115"
```

### Configuration

**NOTE:** The exporter fetches information from a Docker daemon on every scrape, therefore having a too short scrape
interval can impose load on Docker daemon. As a last resort, you can scrape this exporter using a dedicated job with its
own scraping interval.

Below is the command line options summary:

```bash
docker_service_exporter --help
```

| Argument                    | Introduced in Version | Description                                           | Default    |
|-----------------------------|-----------------------|-------------------------------------------------------|------------|
| web.listen-address          | 0.1.0                 | Address to listen on for web interface and telemetry. | :9115      |
| web.telemetry-path          | 0.1.0                 | Path under which to expose metrics.                   | /metrics   |
| docker.service-extra-labels | 0.1.0                 | Additional labels which will be exported.             | []string{} |
| version                     | 0.1.0                 | Show version info on stdout and exit.                 |            |

Commandline parameters are specified with `--`.

### Metrics

| Name                                    | Type    | Cardinality | Help                                  |
|-----------------------------------------|---------|-------------|---------------------------------------|
| dockerservices_meta_created             | counter | 1           | Unix time with milliseconds precision |
| dockerservices_meta_updated             | counter | 1           | Unix time with milliseconds precision |
| dockerservices_spec_replicas            | gauge   | 1           | Required number of service replicas   |
| dockerservices_service_status_running   | gauge   | 0,1         | Actually running services             |
| dockerservices_service_status_desired   | gauge   | 0,1         | Desired number of tasks               |
| dockerservices_service_status_completed | gauge   | 0,1         | Completed number of tasks             |

## Contributing

We welcome any contributions. Please fork the project on GitHub and open
Pull Requests for any proposed changes.

Please note that we will not merge any changes that encourage insecure
behaviour. If in doubt please open an Issue first to discuss your proposal.
