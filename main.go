package main

import (
	"net/http"
	// "github.com/prometheus/client_golang/prometheus"
	// "github.com/prometheus/client_golang/prometheus/promauto"
	"justthetalk/monitoring/jobs"

	"flag"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

func main() {

	pidFile, err := os.Create("/var/run/prom-website-exporter.pid")
	if err != nil {
		log.Error("Unable to create PID file")
		return
	}

	defer func() {
		if r := recover(); r != nil {
			err := r.(error)
			log.Errorf("Fatal error: %v", err)
			debug.PrintStack()
		}
		pidFile.Close()
		os.Remove(pidFile.Name())
	}()

	log.Info("Starting metrics engine")

	configPtr := flag.String("config", "./config.yml", "Path to config file, defaults to ./config.yml")
	flag.Parse()

	log.Infof("Loading config from: %s", *configPtr)

	data, err := ioutil.ReadFile(*configPtr)
	if err != nil {
		panic(err)
	}

	var config jobs.JobConfigFile
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		panic(err)
	}

	metricsEngine := jobs.NewMetricsEngine(&config)
	metricsEngine.Run()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9090", nil)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Info("Stopping metrics engine")
	metricsEngine.Stop()

	log.Info("Exiting")

}
