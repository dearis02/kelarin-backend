package repository

import "github.com/jmoiron/sqlx"

type User interface{}

type userImpl struct {
	db *sqlx.DB
}

func NewUser(db *sqlx.DB) User {
	return &userImpl{db: db}
}
