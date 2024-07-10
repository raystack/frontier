package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

func Init() {
	initStripe()
}

type HistogramFunc func(labelValue ...string) func()

func createMeasureTime(prometheusMetric *prometheus.HistogramVec) HistogramFunc {
	return func(labelValue ...string) func() {
		start := time.Now()
		return func() {
			prometheusMetric.WithLabelValues(labelValue...).Observe(time.Since(start).Seconds())
		}
	}
}
