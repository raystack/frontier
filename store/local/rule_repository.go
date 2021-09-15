package local

import (
	"github.com/ghodss/yaml"
	"github.com/odpf/shield/store"
	"github.com/odpf/shield/structs"
	"github.com/spf13/afero"
	"os"
)

type RuleRepository struct {
	fs    afero.Fs
}

func (repo *RuleRepository) GetAll() ([]structs.Service, error) {
	var services []structs.Service

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
		services = append(services, rule)
	}
	return services, nil
}

func (repo *RuleRepository) readRule(itemName string) (structs.Service, error) {
	fileBytes, err := afero.ReadFile(repo.fs, itemName)
	if err != nil {
		if os.IsNotExist(err) {
			err = store.ErrResourceNotFound
		}
		return structs.Service{}, err
	}
	var s structs.Service
	if err := yaml.Unmarshal(fileBytes, &s); err != nil {
		return structs.Service{}, err
	}
	return s, err
}

func NewRuleRepository(fs afero.Fs) *RuleRepository {
	return &RuleRepository{
		fs: fs,
	}
}