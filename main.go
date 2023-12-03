package main

import (
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/raystack/frontier/cmd"

	_ "github.com/authzed/authzed-go/proto/authzed/api/v0"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/prometheus"
)

func main() {
	cliConfig, err := cmd.LoadConfig()
	if err != nil {
		cliConfig = &cmd.Config{}
	}

	// The exporter embeds a default OpenTelemetry Reader and
	// implements prometheus.Collector, allowing it to be used as
	// both a Reader and Collector.
	exporter, err := prometheus.New()
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	otel.SetMeterProvider(provider)

	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second * 5))
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}

	if err := cmd.New(cliConfig).Execute(); err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
