package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var StripeAPILatency HistogramFunc

func initStripe() {
	StripeAPILatency = createMeasureTime(stripeAPILatencyFactory("api"))
}

var stripeAPILatencyFactory = func(name string) *prometheus.HistogramVec {
	return promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    fmt.Sprintf("stripe_latency_%s", name),
		Help:    "Time took to execute Stripe related API calls",
		Buckets: prometheus.DefBuckets,
	}, []string{"operation", "method"})
}
