package collector

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	dockerClient "github.com/docker/docker/client"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
)

var (
	defaultExportedServiceLabels = []string{
		"com.docker.stack.image",
		"com.docker.stack.namespace",
	}
	defaultServiceLabels      = append([]string{"id", "name", "task_runtime"}, unifyServiceLabels(defaultExportedServiceLabels)...)
	defaultServiceLabelValues = func(s Services, service swarm.Service) []string {
		labels := []string{
			service.ID,
			service.Spec.Annotations.Name,
			string(service.Spec.TaskTemplate.Runtime),
		}

		for _, labelKey := range defaultExportedServiceLabels {
			if labelValue, ok := service.Spec.Annotations.Labels[labelKey]; ok {
				labels = append(labels, labelValue)
			} else {
				labels = append(labels, "")
			}
		}

		for _, labelKey := range s.extraLabels {
			if labelValue, ok := service.Spec.Annotations.Labels[labelKey]; ok {
				labels = append(labels, labelValue)
			} else {
				labels = append(labels, "")
			}
		}

		return labels
	}
)

func unifyServiceLabels(labels []string) []string {
	returnLabels := make([]string, len(labels))

	for labelKey, returnValue := range labels {
		returnLabels[labelKey] = "docker_label_" + strings.Replace(returnValue, ".", "_", -1)
	}
	return returnLabels
}

type serviceMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(node swarm.Service) float64
	Labels func(s Services, service swarm.Service) []string
}

// Services information struct
type Services struct {
	logger         log.Logger
	dockerCli      *dockerClient.Client
	ctx            context.Context
	extraLabels    []string
	up             prometheus.Gauge
	totalScrapes   prometheus.Counter
	serviceMetrics []*serviceMetric
}

func NewServices(logger log.Logger, dockerCli *dockerClient.Client, extraLabels []string, ctx context.Context) *Services {
	serviceLabels := append(defaultServiceLabels, unifyServiceLabels(extraLabels)...)

	return &Services{
		logger:      logger,
		dockerCli:   dockerCli,
		extraLabels: extraLabels,
		ctx:         ctx,
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
					serviceLabels, nil,
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
					serviceLabels, nil,
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
					serviceLabels, nil,
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
					serviceLabels, nil,
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
					serviceLabels, nil,
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
					serviceLabels, nil,
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
				serviceMetric.Labels(s, service)...,
			)
		}
	}
}
