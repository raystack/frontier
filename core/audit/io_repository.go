package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

type WriteOnlyRepository struct {
	writer io.Writer
}

func NewWriteOnlyRepository(writer io.Writer) *WriteOnlyRepository {
	return &WriteOnlyRepository{writer: writer}
}

func (r WriteOnlyRepository) Create(ctx context.Context, log *Log) error {
	err := json.NewEncoder(r.writer).Encode(log)
	return err
}

func (r WriteOnlyRepository) List(ctx context.Context, filter Filter) ([]Log, error) {
	return nil, fmt.Errorf("unsupported")
}

func (r WriteOnlyRepository) GetByID(ctx context.Context, s string) (Log, error) {
	return Log{}, fmt.Errorf("unsupported")
}
