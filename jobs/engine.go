package jobs

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/jinzhu/copier"
	//log "github.com/sirupsen/logrus"
)

type MetricsEngine struct {
	joblist       []Job
	httpDuration  *prometheus.HistogramVec
	httpStatus    *prometheus.CounterVec
	internalError *prometheus.CounterVec
}

func NewMetricsEngine(config *JobConfigFile) *MetricsEngine {

	engine := &MetricsEngine{}

	for _, jobConfig := range config.Jobs {

		var resolvedConfig JobConfig
		copier.CopyWithOption(&resolvedConfig, &config.GlobalConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true})
		copier.CopyWithOption(&resolvedConfig, &jobConfig, copier.Option{IgnoreEmpty: true, DeepCopy: true})

		job := NewBasicJob(&resolvedConfig, engine)

		engine.joblist = append(engine.joblist, job)

	}

	engine.httpDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "http_duration_seconds",
		Help: "Duration of HTTP requests.",
	}, []string{"job", "target"})

	engine.httpStatus = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "http_request_count",
		Help: "Count of HTTP requests",
	}, []string{"job", "target", "status"})

	engine.internalError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "job_error_count",
		Help: "Count of errors running test",
	}, []string{"job", "target"})

	return engine

}

func (engine *MetricsEngine) Run() {
	for _, job := range engine.joblist {
		go job.Run()
	}
}

func (engine *MetricsEngine) Stop() {
	for _, job := range engine.joblist {
		job.Stop()
	}
}

func (engine *MetricsEngine) HttpDuration() *prometheus.HistogramVec {
	return engine.httpDuration
}

func (engine *MetricsEngine) HttpStatus() *prometheus.CounterVec {
	return engine.httpStatus
}

func (engine *MetricsEngine) InternalError() *prometheus.CounterVec {
	return engine.internalError
}
