package postgres

import "github.com/jmoiron/sqlx"

type Store struct {
	DB *sqlx.DB
}

func NewStore(db *sqlx.DB) Store {
	return Store{
		DB: db,
	}
}
