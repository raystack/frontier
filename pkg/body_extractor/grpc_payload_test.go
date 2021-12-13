package body_extractor

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"testing"
)

func TestNewQuery(t *testing.T) {

	t.Run("error", func(t *testing.T) {
		parsedQuery, err := ParseQuery("1.2.3[a].45")
		assert.IsType(t, &strconv.NumError{}, err)

		assert.EqualValues(t, []Query{}, parsedQuery)
	})

	t.Run("success", func(t *testing.T) {
		parsedQuery, err := ParseQuery("1.2.3[1].45")
		assert.Nil(t, err)

		assert.EqualValues(t, []Query{
			{
				Field:    1,
				Index:    -1,
				DataType: Message,
			},
			{
				Field:    2,
				Index:    -1,
				DataType: Message,
			},
			{
				Field:    3,
				Index:    1,
				DataType: NestedArray,
			},
			{
				Field:    45,
				Index:    -1,
				DataType: String,
			},
		}, parsedQuery)
	})
}
