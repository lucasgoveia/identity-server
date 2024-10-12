package respawn

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/labstack/gommon/log"
)

type PostgresRespawner struct {
	schemas []string
}

func NewPostgresRespawner(schemas []string) *PostgresRespawner {
	return &PostgresRespawner{schemas: schemas}
}

func (r *PostgresRespawner) Respawn(connString string) error {
	ctx := context.Background()
	db, err := sql.Open("postgres", connString)

	if err != nil {
		log.Errorf("failed to open connection to db: %s", err)
		return err
	}

	defer func() {
		_ = db.Close()
	}()

	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		log.Errorf("failed to open transaction to db")
		return err
	}

	defer func() {
		if err != nil {
			err := tx.Rollback()
			if err != nil {
				err = fmt.Errorf("transaction rollback failed: %v", err)
			}
			err = fmt.Errorf("transaction rolled back due to error: %v", err)
		}
		err = tx.Commit()
		if err != nil {
			err = errors.New(fmt.Sprintf("failed to commit transaction: %v", err))
		}
	}()

	for _, schema := range r.schemas {

		query := fmt.Sprintf(`
			DO $$ 
			DECLARE
				r RECORD;
			BEGIN
				FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = '%s') LOOP
					-- Truncate each table and reset the identity
					EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' RESTART IDENTITY CASCADE';
				END LOOP;
			END $$;
	`, schema)

		_, err = tx.Exec(query)
		if err != nil {
			return fmt.Errorf("error truncating tables in schema :%s: %v")
		}
	}

	log.Printf("Successfully truncated tables")
	return nil
}
