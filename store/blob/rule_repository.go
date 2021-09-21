package blob

import (
	"context"
	"io"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"
	"github.com/pkg/errors"
	"gocloud.dev/blob"
)

type RuleRepository struct {
	bucket store.Bucket
}

func (repo *RuleRepository) GetAll(ctx context.Context) ([]structs.Ruleset, error) {
	var ruleset []structs.Ruleset

	// get all items
	it := repo.bucket.List(&blob.ListOptions{})
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
		fileBytes, err := repo.bucket.ReadAll(ctx, obj.Key)
		if err != nil {
			return nil, errors.Wrap(err, "bucket.ReadAll: "+obj.Key)
		}

		var s structs.Ruleset
		if err := yaml.Unmarshal(fileBytes, &s); err != nil {
			return nil, errors.Wrap(err, "yaml.Unmarshal: "+obj.Key)
		}
		ruleset = append(ruleset, s)
	}
	return ruleset, nil
}

func NewRuleRepository(b store.Bucket) *RuleRepository {
	return &RuleRepository{
		bucket: b,
	}
}
