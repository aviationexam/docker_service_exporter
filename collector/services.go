package collector

import (
	"github.com/go-kit/log"
	"github.com/prometheus/client_golang/prometheus"
)

// Services information struct
type Services struct {
	logger log.Logger

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter
}

func NewServices(logger log.Logger) *Services {
	return &Services{
		logger: logger,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "node_stats", "up"),
			Help: "Was the last scrape of the Docker services successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "node_stats", "total_scrapes"),
			Help: "Current total Docker services scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "node_stats", "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),
	}
}

func (s Services) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.up.Desc()
	ch <- s.totalScrapes.Desc()
	ch <- s.jsonParseFailures.Desc()
}

func (s Services) Collect(ch chan<- prometheus.Metric) {
	s.totalScrapes.Inc()
	defer func() {
		ch <- s.up
		ch <- s.totalScrapes
		ch <- s.jsonParseFailures
	}()

	//TODO load services
	//dockerServiceResp, err := c.fetchAndDecodeDockerServices()
	//if err != nil {
	//	s.up.Set(0)
	//	_ = level.Warn(c.logger).Log(
	//		"msg", "failed to fetch and decode docker services",
	//		"err", err,
	//	)
	//	return
	//}

	s.up.Set(1)

	// ch <- ...dockerServiceResp
}
