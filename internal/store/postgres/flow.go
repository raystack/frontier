package postgres

import (
	"time"

	"github.com/google/uuid"
	"github.com/odpf/shield/core/authenticate"
)

type Flow struct {
	ID        uuid.UUID `db:"id"`
	Method    string    `db:"method"`
	StartURL  string    `db:"start_url"`
	FinishURL string    `db:"finish_url"`
	Nonce     string    `db:"nonce"`
	CreatedAt time.Time `db:"created_at"`
}

func (f *Flow) transformToFlow() *authenticate.Flow {
	return &authenticate.Flow{
		ID:        f.ID,
		Method:    f.Method,
		StartURL:  f.StartURL,
		FinishURL: f.FinishURL,
		Nonce:     f.Nonce,
		CreatedAt: f.CreatedAt,
	}
}
