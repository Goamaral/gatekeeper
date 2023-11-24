package internal

import (
	"database/sql"
	"os"
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"

	_ "github.com/glebarez/go-sqlite"
)

func NewInjector() *do.Injector {
	injector := do.New()

	do.Provide(injector, func(_ *do.Injector) (*sql.DB, error) {
		// TODO: Implement do.Shutdownable and do.Healthcheckable
		return sql.Open("sqlite", "./db/database.sqlite")
	})

	return injector
}

func NewTestInjector(t *testing.T) *do.Injector {
	injector := NewInjector()

	do.Override(injector, func(_ *do.Injector) (*sql.DB, error) {
		// TODO: Implement do.Shutdownable and do.Healthcheckable
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			return nil, err
		}

		// Load schema
		schemaSqlBytes, err := os.ReadFile(RelativePath("../db/schema.sql"))
		require.NoError(t, err)

		_, err = db.Exec(string(schemaSqlBytes))
		require.NoError(t, err)

		return db, nil
	})

	return injector
}
