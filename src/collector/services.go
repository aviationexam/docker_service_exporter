package collector

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	defaultServiceLabels      = []string{"id", "name", "task_runtime"}
	defaultServiceLabelValues = func(service swarm.Service) []string {
		return []string{
			service.ID,
			service.Spec.Annotations.Name,
			string(service.Spec.TaskTemplate.Runtime),
		}
	}
)

type serviceMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(node swarm.Service) float64
	Labels func(node swarm.Service) []string
}

// Services information struct
type Services struct {
	logger    log.Logger
	dockerCli *dockerClient.Client
	ctx       context.Context

	up             prometheus.Gauge
	totalScrapes   prometheus.Counter
	serviceMetrics []*serviceMetric
}

func NewServices(logger log.Logger, dockerCli *dockerClient.Client, ctx context.Context) *Services {
	return &Services{
		logger:    logger,
		dockerCli: dockerCli,
		ctx:       ctx,
		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, "node_stats", "up"),
			Help: "Was the last scrape of the Docker services successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, "node_stats", "total_scrapes"),
			Help: "Current total Docker services scrapes.",
		}),
		serviceMetrics: []*serviceMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "meta", "created"),
					"Service created at",
					defaultServiceLabels, nil,
				),
				Value: func(service swarm.Service) float64 {
					return float64(service.Meta.CreatedAt.UnixMilli()) / 1000
				},
				Labels: defaultServiceLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "meta", "updated"),
					"Service updated at",
					defaultServiceLabels, nil,
				),
				Value: func(service swarm.Service) float64 {
					return float64(service.Meta.UpdatedAt.UnixMilli()) / 1000
				},
				Labels: defaultServiceLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "spec", "replicas"),
					"Service replicas",
					defaultServiceLabels, nil,
				),
				Value: func(service swarm.Service) float64 {
					if service.Spec.Mode.Replicated != nil {
						return float64(*service.Spec.Mode.Replicated.Replicas)
					}

					return -1
				},
				Labels: defaultServiceLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "service_status", "running"),
					"Service replicas",
					defaultServiceLabels, nil,
				),
				Value: func(service swarm.Service) float64 {
					return float64(service.ServiceStatus.RunningTasks)
				},
				Labels: defaultServiceLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "service_status", "desired"),
					"Service replicas",
					defaultServiceLabels, nil,
				),
				Value: func(service swarm.Service) float64 {
					return float64(service.ServiceStatus.DesiredTasks)
				},
				Labels: defaultServiceLabelValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "service_status", "completed"),
					"Service replicas",
					defaultServiceLabels, nil,
				),
				Value: func(service swarm.Service) float64 {
					return float64(service.ServiceStatus.CompletedTasks)
				},
				Labels: defaultServiceLabelValues,
			},
		},
	}
}

func (s Services) Describe(ch chan<- *prometheus.Desc) {
	for _, serviceMetric := range s.serviceMetrics {
		ch <- serviceMetric.Desc
	}

	ch <- s.up.Desc()
	ch <- s.totalScrapes.Desc()
}

func (s Services) Collect(ch chan<- prometheus.Metric) {
	s.totalScrapes.Inc()
	defer func() {
		ch <- s.up
		ch <- s.totalScrapes
	}()

	services, err := s.dockerCli.ServiceList(s.ctx, types.ServiceListOptions{
		Status: true,
	})
	if err != nil {
		s.up.Set(0)
		_ = level.Warn(s.logger).Log(
			"msg", "failed to fetch docker services",
			"err", err,
		)
		return
	}

	s.up.Set(1)

	for serviceId := range services {
		service := services[serviceId]

		for _, serviceMetric := range s.serviceMetrics {
			ch <- prometheus.MustNewConstMetric(
				serviceMetric.Desc,
				serviceMetric.Type,
				serviceMetric.Value(service),
				serviceMetric.Labels(service)...,
			)
		}
	}
}
