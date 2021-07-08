package jobs

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"

	"net/http"
	"runtime/debug"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

type JobConfigTarget struct {
	Targets []string `yaml:"targets"`
}

type JobConfig struct {
	JobName        string            `yaml:"job_name"`
	ScrapeInterval string            `yaml:"scrape_interval"`
	ExternalLabels map[string]string `yaml:"external_labels"`
	StaticConfigs  []JobConfigTarget `yaml:"static_configs"`
}

type JobConfigFile struct {
	GlobalConfig JobConfig   `yaml:"global"`
	Jobs         []JobConfig `yaml:"jobs"`
}

type Job interface {
	Run()
	Stop()
}

type BasicJob struct {
	config   *JobConfig
	engine   *MetricsEngine
	quitWait sync.WaitGroup
	quitFlag chan bool
}

func NewBasicJob(config *JobConfig, engine *MetricsEngine) Job {

	job := &BasicJob{
		engine:   engine,
		config:   config,
		quitFlag: make(chan bool),
	}

	return job

}

func (job *BasicJob) Run() {

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("Fatal error: %v", err)
			debug.PrintStack()
		}
	}()

	log.Infof("Running: %s", job.config.JobName)

	job.quitWait.Add(1)
	defer job.quitWait.Done()

	period, err := time.ParseDuration(job.config.ScrapeInterval)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(period)
	defer ticker.Stop()

	quit := false
	for !quit {
		select {
		case <-ticker.C:
			job.execute()
		case <-job.quitFlag:
			quit = true
		}
	}

}

func (job *BasicJob) Stop() {
	job.quitFlag <- true
	job.quitWait.Wait()
}

func (job *BasicJob) execute() {

	log.Infof("Executing: %s", job.config.JobName)

	for _, targets := range job.config.StaticConfigs {
		for _, target := range targets.Targets {

			timer := prometheus.NewTimer(job.engine.HttpDuration().WithLabelValues(job.config.JobName, target))
			statusCode := job.makeRequest(target)
			timer.ObserveDuration()

			job.engine.HttpStatus().WithLabelValues(job.config.JobName, target, strconv.Itoa(statusCode)).Inc()

		}
	}

}

func (job *BasicJob) makeRequest(target string) int {

	defer func() {
		job.engine.InternalError().WithLabelValues(job.config.JobName, target).Inc()
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("Fatal error: %v", err)
			debug.PrintStack()
		}
	}()

	resp, err := http.Get(target)
	if err != nil {
		panic(err)
	}

	return resp.StatusCode

}
