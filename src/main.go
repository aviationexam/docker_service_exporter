package main

import (
	"github.com/aviationexam/docker_service_exporter/collector"
	"net/http"
	"os"
	"os/signal"
	"time"

	"context"

	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	dockerClient "github.com/docker/docker/client"
)

const name = "docker_service_exporter"

func main() {
	var (
		listenAddress = kingpin.Flag(
			"web.listen-address",
			"Address to listen on for web interface and telemetry.",
		).Default(":9115").String()
		metricsPath = kingpin.Flag(
			"web.telemetry-path",
			"Path under which to expose metrics.",
		).Default("/metrics").String()
		logLevel = kingpin.Flag(
			"log.level",
			"Sets the loglevel. Valid levels are '"+logLevelDebug+"', '"+logLevelInfo+"', '"+logLevelWarn+"', '"+loglevelError+"'",
		).Default(logLevelInfo).String()
		logFormat = kingpin.Flag(
			"log.format",
			"Sets the log format. Valid formats are '"+logFmtJson+"' and '"+logFmtFmt+"'",
		).Default(
			logFmtFmt,
		).String()
		logOutput = kingpin.Flag(
			"log.output",
			"Sets the log output. Valid outputs are '"+logOutputStderr+"' and '"+logOutputStdout+"'",
		).Default(logOutputStderr).String()
	)

	kingpin.Version(version.Print(name))
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	logger := getLogger(*logLevel, *logOutput, *logFormat)

	// Create a context that is cancelled on SIGKILL or SIGINT.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	dockerCli, err := dockerClient.NewClientWithOpts(dockerClient.FromEnv)
	if err != nil {
		_ = level.Error(logger).Log(
			"msg", "failed create docker client",
			"err", err,
		)
		os.Exit(1)
	}

	// version metric
	prometheus.MustRegister(version.NewCollector(name))

	prometheus.MustRegister(collector.NewServices(logger, dockerCli, ctx))

	// create a http server
	server := &http.Server{}

	mux := http.DefaultServeMux
	mux.Handle(*metricsPath, promhttp.Handler())
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte(`<html>
			<head><title>Docker Service Exporter</title></head>
			<body>
			<h1>Docker Service Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			<p><a href="/healthz">Health check</a></p>
			</body>
			</html>`,
		))
		if err != nil {
			_ = level.Error(logger).Log(
				"msg", "failed handling writer",
				"err", err,
			)
		}
	})

	// health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
	})

	server.Handler = mux
	server.Addr = *listenAddress

	_ = level.Info(logger).Log(
		"msg", "starting docker_service_exporter",
		"addr", *listenAddress,
	)

	go func() {
		if err := server.ListenAndServe(); err != nil {
			_ = level.Error(logger).Log(
				"msg", "http server quit",
				"err", err,
			)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	_ = level.Info(logger).Log("msg", "shutting down")
	// create a context for graceful http server shutdown
	srvCtx, srvCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer srvCancel()
	_ = server.Shutdown(srvCtx)
}
