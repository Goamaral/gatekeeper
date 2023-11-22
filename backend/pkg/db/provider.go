package db

import (
	"github.com/jmoiron/sqlx"
)

type Provider struct {
	DB *sqlx.DB
}

func NewProvider(driverName, dataSourceName string) (Provider, error) {
	db, err := sqlx.Open(driverName, dataSourceName)
	if err != nil {
		return Provider{}, err
	}
	return Provider{DB: db}, nil
}

func (p Provider) Close() error {
	return p.DB.Close()
}
