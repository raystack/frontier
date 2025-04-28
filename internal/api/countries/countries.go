package countries

import (
	"encoding/json"
	"net/http"

	"github.com/raystack/frontier/internal/store/countries"
)

func ListCountriesHandler(w http.ResponseWriter, _r *http.Request) {
	countries, err := countries.ListCountries()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := json.NewEncoder(w).Encode(countries); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
