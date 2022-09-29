package schema

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendIfUnique(t *testing.T) {
	fmt.Println(AppendIfUnique([]string{"1", "2", "3"}, []string{"3", "4"}))
	assert.ElementsMatch(t, AppendIfUnique([]string{"1", "2", "3"}, []string{"3", "4"}), []string{"1", "2", "3", "4"})
}
