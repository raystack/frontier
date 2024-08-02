package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var DatabaseQueryLatency HistogramFunc

func initDB() {
	DatabaseQueryLatency = createMeasureTime(promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "db_query_latency",
		Help:    "Time took to execute database query",
		Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 30, 60},
	}, []string{"collection", "operation"}))
}
