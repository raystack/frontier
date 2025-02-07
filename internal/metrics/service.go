package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var ServiceOprLatency HistogramFunc

func initService() {
	ServiceOprLatency = createMeasureTime(promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "service_operation_latency",
		Help:    "Time taken for service operation to complete",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
	}, []string{"service", "operation"}))
}
