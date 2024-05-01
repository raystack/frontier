package audit

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestStructPB(t *testing.T) {
	input := make(map[string]interface{})
	input["key"] = "value"
	input["data"] = map[string]interface{}{
		"key2": "value2",
	}

	result, err := structpb.NewStruct(input)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// map of string fails
	input["data2"] = map[string]string{
		"key3": "value3",
	}
	_, err = structpb.NewStruct(input)
	assert.Error(t, err)
	delete(input, "data2")

	now := time.Now()
	logDecoded := map[string]interface{}{}
	err = mapstructure.Decode(&Log{
		Source: "source",
		Target: Target{
			ID:   "target-id",
			Type: "target-type",
		},
		Actor: Actor{
			ID:   "actor-id",
			Type: "actor-type",
			Name: "actor-name",
		},
		Metadata:  map[string]string{},
		Action:    "action",
		ID:        "id",
		CreatedAt: now,
	}, &logDecoded)
	assert.NoError(t, err)
	assert.NotNil(t, logDecoded)
}

func TestTransformToEventData(t *testing.T) {
	now := time.Now()
	type args struct {
		l *Log
	}
	tests := []struct {
		name string
		args args
		want map[string]interface{}
	}{
		{
			name: "should decode everything except metadata",
			args: args{
				l: &Log{
					Source: "source",
					Target: Target{
						ID:   "target-id",
						Type: "target-type",
					},
					Actor: Actor{
						ID:   "actor-id",
						Type: "actor-type",
						Name: "actor-name",
					},
					Metadata:  map[string]string{},
					Action:    "action",
					ID:        "id",
					CreatedAt: now,
				},
			},
			want: map[string]interface{}{
				"source":   "source",
				"target":   map[string]any{"id": "target-id", "type": "target-type"},
				"actor":    map[string]any{"id": "actor-id", "type": "actor-type", "name": "actor-name"},
				"metadata": map[string]any{},
			},
		},
		{
			name: "should decode metadata correctly",
			args: args{
				l: &Log{
					Source: "source",
					Metadata: map[string]string{
						"key": "value",
					},
				},
			},
			want: map[string]interface{}{
				"source": "source",
				"actor":  map[string]any{},
				"target": map[string]any{},
				"metadata": map[string]any{
					"key": "value",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TransformToEventData(tt.args.l)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("TransformToEventData() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
