package body_extractor

import (
	"strconv"
	"testing"

	fixturesv1 "github.com/odpf/shield/pkg/body_extractor/fixtures"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
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
				DataType: MessageArray,
			},
			{
				Field:    45,
				Index:    -1,
				DataType: String,
			},
		}, parsedQuery)
	})
}

func TestExtract(t *testing.T) {
	testMessage := fixturesv1.NestedMessageL0{
		L1: &fixturesv1.NestedMessageL1{
			L2: &fixturesv1.NestedMessageL2{
				L3: []*fixturesv1.NestedMessageL3{
					{S1L3: "s1l3_one"},
					{S1L3: "s1l3_two"},
					{S1L3: "s1l3_three"},
				},
			},
		},
	}

	testgrpcPayloadHandler := GRPCPayloadHandler{grpcDisabled: true}

	msg, err := proto.Marshal(&testMessage)

	assert.NoError(t, err)

	ex, err := testgrpcPayloadHandler.extractFromRequest(msg, "1.2.7[1]")
	assert.NoError(t, err)
	assert.EqualValues(t, []string{
		"\n\bs1l3_one",
		"\n\bs1l3_two",
		"\n\ns1l3_three",
	}, ex.([]string))
}
