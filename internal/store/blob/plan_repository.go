package blob

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/raystack/frontier/billing/feature"
	"github.com/raystack/frontier/billing/plan"

	"gocloud.dev/blob"
	"gopkg.in/yaml.v3"
)

type PlanRepository struct {
	bucket Bucket
}

func NewPlanRepository(b Bucket) *PlanRepository {
	return &PlanRepository{bucket: b}
}

// Get returns the plans from the bucket
func (s *PlanRepository) Get(ctx context.Context) (plan.File, error) {
	var definitions []plan.File

	// iterate over bucket files, only read .yml & .yaml files
	it := s.bucket.List(&blob.ListOptions{})
	for {
		obj, err := it.Next(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return plan.File{}, err
		}

		if obj.IsDir {
			continue
		}
		if !(strings.HasSuffix(obj.Key, ".yaml") || strings.HasSuffix(obj.Key, ".yml")) {
			continue
		}
		fileBytes, err := s.bucket.ReadAll(ctx, obj.Key)
		if err != nil {
			return plan.File{}, fmt.Errorf("%s: %s", "error in reading bucket object", err.Error())
		}

		var def plan.File
		if err := yaml.Unmarshal(fileBytes, &def); err != nil {
			return plan.File{}, fmt.Errorf("get: yaml.Unmarshal: %s: %w", obj.Key, err)
		}
		definitions = append(definitions, def)
	}

	var allPlans []plan.Plan
	var allFeatures []feature.Feature
	for _, definition := range definitions {
		allPlans = append(allPlans, definition.Plans...)
		allFeatures = append(allFeatures, definition.Features...)
	}
	return plan.File{
		Plans:    allPlans,
		Features: allFeatures,
	}, nil
}
