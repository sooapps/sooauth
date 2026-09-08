package migrate

import (
	"database/sql"
	"fmt"
	"io/fs"

	"github.com/pressly/goose/v3"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/sooapps/sooauth/server/migrations"
)

func Up(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, ".")
}

func Status(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Status(db, ".")
}

func SubFS() fs.FS {
	return migrations.FS
}
