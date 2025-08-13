package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics
var (
	podsProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "kustom_controller_pods_processed_total",
		Help: "Total number of pods processed by the controller",
	}, []string{"namespace", "type"})

	resourcesEnforced = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kustom_controller_resources_enforced_total",
		Help: "Total number of resource enforcements performed",
	})

	processingTime = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "kustom_controller_processing_duration_seconds",
		Help:    "Time taken to process pod events",
		Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5},
	})

	errorsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kustom_controller_errors_total",
		Help: "Total number of errors encountered",
	})
)
