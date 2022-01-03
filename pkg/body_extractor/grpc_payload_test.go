package body_extractor

import (
	"bytes"
	"io/ioutil"
	"testing"

	fixturesv1 "github.com/odpf/shield/pkg/body_extractor/fixtures"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestNewQuery(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		parsedQuery, err := ParseQuery("1.2.3a.45")
		assert.NotNil(t, err)

		assert.Nil(t, parsedQuery)
	})

	t.Run("root string", func(t *testing.T) {
		parsedQuery, err := ParseQuery("1")
		assert.Nil(t, err)

		assert.EqualValues(t, []Query{{
			Field:    1,
			DataType: String,
		}}, parsedQuery)
	})

	t.Run("root repeated string", func(t *testing.T) {
		parsedQuery, err := ParseQuery("1[*]")
		assert.Nil(t, err)

		assert.EqualValues(t, []Query{{
			Field:    1,
			DataType: StringArray,
		}}, parsedQuery)
	})

	t.Run("root repeated message", func(t *testing.T) {
		parsedQuery, err := ParseQuery("1[*].2")
		assert.Nil(t, err)

		assert.EqualValues(t, []Query{
			{
				Field:    1,
				DataType: MessageArray,
			},
			{
				Field:    2,
				DataType: String},
		}, parsedQuery)
	})

	t.Run("repeated message nested", func(t *testing.T) {
		parsedQuery, err := ParseQuery("1.2.3[*].45")
		assert.Nil(t, err)

		assert.EqualValues(t, []Query{
			{
				Field:    1,
				DataType: Message,
			},
			{
				Field:    2,
				DataType: Message,
			},
			{
				Field:    3,
				DataType: MessageArray,
			},
			{
				Field:    45,
				DataType: String,
			},
		}, parsedQuery)
	})

	t.Run("repeated message nested with repeated string", func(t *testing.T) {
		parsedQuery, err := ParseQuery("1.2.3[*].45.8[*]")
		assert.Nil(t, err)

		assert.EqualValues(t, []Query{
			{
				Field:    1,
				DataType: Message,
			},
			{
				Field:    2,
				DataType: Message,
			},
			{
				Field:    3,
				DataType: MessageArray,
			},
			{
				Field:    45,
				DataType: Message,
			},
			{
				Field:    8,
				DataType: StringArray,
			},
		}, parsedQuery)
	})
}

func TestExtract(t *testing.T) {
	t.Parallel()
	table := []struct {
		title       string
		testMessage proto.Message
		query       string
		want        interface{}
		err         error
	}{
		{
			title:       "root string",
			testMessage: &fixturesv1.NestedMessageL3{S1L3: "S1L3"},
			query:       "1",
			want:        "S1L3",
			err:         nil,
		},
		{
			title:       "message string",
			testMessage: &fixturesv1.NestedMessageL3{L5: &fixturesv1.NestedMessageL5{S1L5: "SomeMessage"}},
			query:       "9.12",
			want:        "SomeMessage",
			err:         nil,
		},
		{
			title: "root repeated",
			testMessage: &fixturesv1.NestedMessageL4{
				S3L4: []string{
					"two",
					"four",
					"six",
					"ten",
				},
			},
			query: "1.2.7[*]",
			want: []interface{}{
				"two",
				"four",
				"six",
				"ten",
			},
		},
		{
			title: "message with repeated string",
			testMessage: &fixturesv1.NestedMessageL0{
				L1: &fixturesv1.NestedMessageL1{
					L2: &fixturesv1.NestedMessageL2{
						L3: []*fixturesv1.NestedMessageL3{
							{S1L3: "s1l3_one"},
							{S1L3: "s1l3_two"},
							{S1L3: "s1l3_three"},
						},
					},
				},
			},
			query: "1.2.7[*]",
			want: []interface{}{
				"\n\bs1l3_one",
				"\n\bs1l3_two",
				"\n\ns1l3_three",
			},
		},
		{
			title: "message with repeated message",
			testMessage: &fixturesv1.NestedMessageL0{
				L1: &fixturesv1.NestedMessageL1{
					L2: &fixturesv1.NestedMessageL2{
						L3: []*fixturesv1.NestedMessageL3{
							{S1L3: "s1l3_one"},
							{S1L3: "s1l3_two"},
							{S1L3: "s1l3_three"},
						},
						L4: []*fixturesv1.NestedMessageL4{
							{
								S1L4: "S1L4_one",
							},
							{
								S1L4: "S1L4_two",
							},
							{
								S1L4: "S1L4_three",
							},
							{
								S1L4: "S1L4_four",
							},
						},
					},
				},
			},
			query: "1.2.8[*].1",
			want: []interface{}{
				"S1L4_one",
				"S1L4_two",
				"S1L4_three",
				"S1L4_four",
			},
		},
		{
			title: "message with repeated message having repeated string",
			testMessage: &fixturesv1.NestedMessageL0{
				L1: &fixturesv1.NestedMessageL1{
					L2: &fixturesv1.NestedMessageL2{
						L3: []*fixturesv1.NestedMessageL3{
							{S1L3: "s1l3_one"},
							{S1L3: "s1l3_two"},
							{S1L3: "s1l3_three"},
						},
						L4: []*fixturesv1.NestedMessageL4{
							{
								S1L4: "S1L4_one",
								S3L4: []string{
									"S3L4_one_one",
									"S3L4_one_two",
									"S3L4_one_three",
									"S3L4_one_four",
								},
							},
							{
								S1L4: "S1L4_two",
								S3L4: []string{
									"S3L4_two_one",
									"S3L4_two_two",
									"S3L4_two_three",
								},
							},
							{
								S1L4: "S1L4_three",
								S3L4: []string{
									"S3L4_three_one",
									"S3L4_three_two",
								},
							},
							{
								S1L4: "S1L4_four",
								S3L4: []string{
									"S3L4_four_one",
								},
							},
						},
					},
				},
			},
			query: "1.2.8[*].3[*]",
			want: []interface{}{
				"S3L4_one_one",
				"S3L4_one_two",
				"S3L4_one_three",
				"S3L4_one_four",
				"S3L4_two_one",
				"S3L4_two_two",
				"S3L4_two_three",
				"S3L4_three_one",
				"S3L4_three_two",
				"S3L4_four_one",
			},
		},
	}

	testgrpcPayloadHandler := GRPCPayloadHandler{grpcDisabled: true}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			t.Parallel()
			msg, err := proto.Marshal(tt.testMessage)
			assert.NoError(t, err)

			testReader := ioutil.NopCloser(bytes.NewBuffer(msg))
			extractedData, err := testgrpcPayloadHandler.Extract(&testReader, tt.query)

			assert.EqualValues(t, tt.want, extractedData)
			assert.Equal(t, err, tt.err)
		})
	}
}
