package postgres

import (
	"fmt"
	"testing"
)

func TestBuildGetUserQuery(t *testing.T) {
	goquQuery, err := buildGetUserQuery(dialect)
	if err != nil {
		fmt.Errorf("%s", err)
	}

	fmt.Println(goquQuery)
}

// Pagination
var page uint = 2
var limit uint = 20
var offset = (page - 1) * limit

// Filtration
emailBit := "nihar"
nameBit := "niha"