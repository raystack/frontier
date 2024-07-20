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
		Buckets: prometheus.DefBuckets,
	}, []string{"collection", "operation"}))
}
