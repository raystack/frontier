package countries

import (
	"embed"
	"encoding/json"
)

//go:embed all:countries.json
var Assets embed.FS

type Country struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

func ListCountries() ([]Country, error) {
	data, err := Assets.ReadFile("countries.json")
	if err != nil {
		return nil, err
	}

	var countries []Country
	err = json.Unmarshal(data, &countries)
	if err != nil {
		return nil, err
	}

	return countries, nil
}
