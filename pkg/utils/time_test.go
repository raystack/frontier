package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConvertToStartOfDay(t *testing.T) {
	input := time.Date(2020, time.April, 11, 21, 34, 01, 0, time.UTC)
	startOfDayTime := ConvertToStartOfDay(input)

	expectedOutput := time.Date(2020, time.April, 11, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, startOfDayTime, expectedOutput)
}
