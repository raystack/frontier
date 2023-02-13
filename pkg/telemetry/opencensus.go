package telemetry

import (
	"context"

	"contrib.go.opencensus.io/exporter/prometheus"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
)

var (
	// The number of new metadata keys
	MMissingMetadataKeys    = stats.Int64("metadata-keys/counter", "The number of missing metadata keys", "1")
	MResourceFailedToCreate = stats.Int64("resource-failed-to-create/counter", "The number of resources failed to be created", "1")
)

var (
	KeyMethod, _                   = tag.NewKey("method")
	KeyMissingKey, _               = tag.NewKey("missing-key")
	KeyResourceCreationResponse, _ = tag.NewKey("resource-creation-response")
)

var (
	MissingMetadataKeysView = &view.View{
		Name:        "metadata-keys/counter",
		Measure:     MMissingMetadataKeys,
		Description: "The number of missing metadata keys",

		Aggregation: view.Count(),
		TagKeys:     []tag.Key{KeyMethod, KeyMissingKey}}

	ResourceFailedToCreateView = &view.View{
		Name:        "resource-failed-to-create/counter",
		Measure:     MResourceFailedToCreate,
		Description: "The number of resources failed to be created",

		Aggregation: view.Count(),
		TagKeys:     []tag.Key{KeyMethod, KeyResourceCreationResponse}}
)

func SetupOpenCensus(ctx context.Context, cfg Config) (*prometheus.Exporter, error) {
	if err := setupViews(); err != nil {
		return nil, err
	}

	pe, err := prometheus.NewExporter(prometheus.Options{
		Namespace: cfg.ServiceName,
	})
	if err != nil {
		return nil, err
	}
	view.RegisterExporter(pe)
	return pe, nil
}

func setupViews() error {
	err := view.Register(MissingMetadataKeysView, ResourceFailedToCreateView)
	if err != nil {
		return err
	}

	return nil
}
