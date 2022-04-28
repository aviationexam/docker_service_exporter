package main

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"io"
	"strings"
)

const logOutputStderr = "stderr"
const logOutputStdout = "stdout"

const logFmtJson = "json"
const logFmtFmt = "logfmt"

const logLevelDebug = "debug"
const logLevelInfo = "info"
const logLevelWarn = "warn"
const loglevelError = "error"

func getLogger(logLevel, logOutput, logFmt string) log.Logger {
	var out *os.File
	switch strings.ToLower(logOutput) {
	case logOutputStderr:
		out = os.Stderr
	case logOutputStdout:
		out = os.Stdout
	default:
		out = os.Stdout
	}

	var logCreator func(io.Writer) log.Logger
	switch strings.ToLower(logFmt) {
	case logFmtJson:
		logCreator = log.NewJSONLogger
	case logFmtFmt:
		logCreator = log.NewLogfmtLogger
	default:
		logCreator = log.NewLogfmtLogger
	}

	// create a logger
	logger := logCreator(log.NewSyncWriter(out))

	// set loglevel
	var loglevelFilterOpt level.Option
	switch strings.ToLower(logLevel) {
	case logLevelDebug:
		loglevelFilterOpt = level.AllowDebug()
	case logLevelInfo:
		loglevelFilterOpt = level.AllowInfo()
	case logLevelWarn:
		loglevelFilterOpt = level.AllowWarn()
	case loglevelError:
		loglevelFilterOpt = level.AllowError()
	default:
		loglevelFilterOpt = level.AllowInfo()
	}

	logger = level.NewFilter(logger, loglevelFilterOpt)
	logger = log.With(
		logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	return logger
}
