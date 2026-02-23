package operator

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Metrics holds all Prometheus metrics exported by the operator.
type Metrics struct {
	PipelineRunsTotal   *prometheus.CounterVec
	PipelineRunsActive  prometheus.Gauge
	DataItemsTotal      *prometheus.CounterVec
	StageDuration       *prometheus.HistogramVec
	ContainersActive    prometheus.Gauge
}

// NewMetrics creates and registers all Prometheus metrics with the default
// registerer. It returns a ready-to-use Metrics instance.
func NewMetrics() *Metrics {
	m := &Metrics{
		PipelineRunsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "portwhine_pipeline_runs_total",
				Help: "Total number of pipeline runs.",
			},
			[]string{"pipeline_id", "status"},
		),
		PipelineRunsActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "portwhine_pipeline_runs_active",
				Help: "Number of currently active pipeline runs.",
			},
		),
		DataItemsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "portwhine_data_items_total",
				Help: "Total number of data items persisted.",
			},
			[]string{"type"},
		),
		StageDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "portwhine_stage_duration_seconds",
				Help:    "Duration of stage execution in seconds.",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"node_id", "status"},
		),
		ContainersActive: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "portwhine_containers_active",
				Help: "Number of currently active containers.",
			},
		),
	}

	prometheus.MustRegister(
		m.PipelineRunsTotal,
		m.PipelineRunsActive,
		m.DataItemsTotal,
		m.StageDuration,
		m.ContainersActive,
	)

	return m
}

// RunStarted implements pipeline.EngineMetrics. It increments the total run
// counter (with a "started" status) and the active runs gauge.
func (m *Metrics) RunStarted(pipelineID string) {
	m.PipelineRunsTotal.WithLabelValues(pipelineID, "started").Inc()
	m.PipelineRunsActive.Inc()
}

// RunFinished implements pipeline.EngineMetrics. It increments the total run
// counter with the final status label and decrements the active runs gauge.
func (m *Metrics) RunFinished(pipelineID, status string) {
	m.PipelineRunsTotal.WithLabelValues(pipelineID, status).Inc()
	m.PipelineRunsActive.Dec()
}

// DataItemPersisted implements pipeline.EngineMetrics. It increments the data
// items counter with the item type label.
func (m *Metrics) DataItemPersisted(itemType string) {
	m.DataItemsTotal.WithLabelValues(itemType).Inc()
}

// StageFinished implements pipeline.EngineMetrics. It records the stage
// execution duration in the histogram.
func (m *Metrics) StageFinished(nodeID, status string, duration time.Duration) {
	m.StageDuration.WithLabelValues(nodeID, status).Observe(duration.Seconds())
}
