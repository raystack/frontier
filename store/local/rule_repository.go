package local

import (
	"errors"
	"os"

	"github.com/odpf/shield/core/rule"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

var (
	ErrResourceNotFound = errors.New("resource not found")
)

// TODO: might delete this
type RuleRepository struct {
	fs afero.Fs
}

func (repo *RuleRepository) GetAll() ([]rule.Ruleset, error) {
	var rulesets []rule.Ruleset

	dirItems, err := afero.ReadDir(repo.fs, ".")
	if err != nil {
		return nil, err
	}
	for _, dirItem := range dirItems {
		if dirItem.IsDir() {
			// we ignore nested directories for now
			continue
		}

		rule, err := repo.readRule(dirItem.Name())
		if err != nil {
			return nil, err
		}
		rulesets = append(rulesets, rule)
	}
	return rulesets, nil
}

func (repo *RuleRepository) readRule(itemName string) (rule.Ruleset, error) {
	fileBytes, err := afero.ReadFile(repo.fs, itemName)
	if err != nil {
		if os.IsNotExist(err) {
			err = ErrResourceNotFound
		}
		return rule.Ruleset{}, err
	}
	var s rule.Ruleset
	if err := yaml.Unmarshal(fileBytes, &s); err != nil {
		return rule.Ruleset{}, err
	}
	return s, err
}

func NewRuleRepository(fs afero.Fs) *RuleRepository {
	return &RuleRepository{
		fs: fs,
	}
}
