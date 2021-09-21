package local

import (
	"os"

	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v3"
)

// TODO: might delete this
type RuleRepository struct {
	fs afero.Fs
}

func (repo *RuleRepository) GetAll() ([]structs.Ruleset, error) {
	var rulesets []structs.Ruleset

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

func (repo *RuleRepository) readRule(itemName string) (structs.Ruleset, error) {
	fileBytes, err := afero.ReadFile(repo.fs, itemName)
	if err != nil {
		if os.IsNotExist(err) {
			err = store.ErrResourceNotFound
		}
		return structs.Ruleset{}, err
	}
	var s structs.Ruleset
	if err := yaml.Unmarshal(fileBytes, &s); err != nil {
		return structs.Ruleset{}, err
	}
	return s, err
}

func NewRuleRepository(fs afero.Fs) *RuleRepository {
	return &RuleRepository{
		fs: fs,
	}
}
