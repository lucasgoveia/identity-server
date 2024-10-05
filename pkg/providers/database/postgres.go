package database

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/uptrace/opentelemetry-go-extra/otelsql"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

type Db struct {
	Db *sql.DB
}

func NewPostgresDb(url string) (*Db, error) {
	db, err := otelsql.Open("postgres", url,
		otelsql.WithAttributes(semconv.DBSystemPostgreSQL),
		otelsql.WithDBName("identity-server-db"))

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
