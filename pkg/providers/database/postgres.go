package database

import (
	"database/sql"
	_ "github.com/lib/pq"
)

type Db struct {
	Db *sql.DB
}

func NewPostgresDb(url string) (*Db, error) {
	db, err := sql.Open("postgres", url)

	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &Db{
		Db: db,
	}, nil
}

func (p *Db) Close() error {
	return p.Db.Close()
}

func (p *Db) GetProviderType() string {
	return "postgres"
}
