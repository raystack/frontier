package blob

import (
	"context"
	"fmt"
	"io"
	"strings"

	"gocloud.dev/blob"
	"gopkg.in/yaml.v3"
)

type PlanFile struct {
	Plans []Plan `json:"plans" yaml:"plans"`
}

type Plan struct {
	Name        string            `json:"name" yaml:"name"`
	Title       string            `json:"title" yaml:"title"`
	Description string            `json:"description" yaml:"description"`
	Features    []Feature         `json:"features" yaml:"features"`
	Metadata    map[string]string `json:"metadata" yaml:"metadata"`
}

type Feature struct {
	Name        string            `json:"name" yaml:"name"`
	Title       string            `json:"title" yaml:"title"`
	Description string            `json:"description" yaml:"description"`
	Prices      []Price           `json:"prices" yaml:"prices"`
	Metadata    map[string]string `json:"metadata" yaml:"metadata"`
}

type Price struct {
	Name              string            `json:"name" yaml:"name"`
	Amount            int64             `json:"amount" yaml:"amount"`
	Currency          string            `json:"currency" yaml:"currency"`
	Type              string            `json:"type" yaml:"type"`
	RecurringInterval string            `json:"recurring_interval" yaml:"recurring_interval"`
	UsageType         string            `json:"usage_type" yaml:"usage_type"`
	MeteredAggregate  string            `json:"metered_aggregate" yaml:"metered_aggregate"`
	Metadata          map[string]string `json:"metadata" yaml:"metadata"`
}

type PlanRepository struct {
	bucket Bucket
}

func NewPlanRepository(b Bucket) *PlanRepository {
	return &PlanRepository{bucket: b}
}

// Get returns the plans from the bucket
func (s *PlanRepository) Get(ctx context.Context) ([]Plan, error) {
	var definitions []PlanFile

	// iterate over bucket files, only read .yml & .yaml files
	it := s.bucket.List(&blob.ListOptions{})
	for {
		obj, err := it.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if obj.IsDir {
			continue
		}
		if !(strings.HasSuffix(obj.Key, ".yaml") || strings.HasSuffix(obj.Key, ".yml")) {
			continue
		}
		fileBytes, err := s.bucket.ReadAll(ctx, obj.Key)
		if err != nil {
			return nil, fmt.Errorf("%s: %s", "error in reading bucket object", err.Error())
		}

		var def PlanFile
		if err := yaml.Unmarshal(fileBytes, &def); err != nil {
			return nil, fmt.Errorf("get: yaml.Unmarshal: %s: %w", obj.Key, err)
		}
		definitions = append(definitions, def)
	}

	allPlans := []Plan{}
	for _, definition := range definitions {
		allPlans = append(allPlans, definition.Plans...)
	}
	return allPlans, nil
}
