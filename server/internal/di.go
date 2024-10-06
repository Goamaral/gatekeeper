package internal

import (
	"database/sql"
	"fmt"
	"gatekeeper/internal/helper"
	"gatekeeper/pkg/fs"
	"gatekeeper/pkg/jwt_provider"
	"os"
	"testing"

	"github.com/samber/do"
	"github.com/stretchr/testify/require"

	_ "github.com/glebarez/go-sqlite"
)

func NewInjector() *do.Injector {
	i := do.New()

	do.Provide(i, func(_ *do.Injector) (*sql.DB, error) {
		// TODO: Implement do.Shutdownable and do.Healthcheckable
		return sql.Open("sqlite", helper.RelativePath("../db/database.sqlite"))
	})

	do.Provide(i, func(i *do.Injector) (jwt_provider.Provider, error) {
		privKeyFile, err := os.Open(fs.RelativePath("../secrets/ecdsa"))
		if err != nil {
			return jwt_provider.Provider{}, fmt.Errorf("failed to open private key file: %w", err)
		}
		pubKeyFile, err := os.Open(fs.RelativePath("../secrets/ecdsa.pub"))
		if err != nil {
			return jwt_provider.Provider{}, fmt.Errorf("failed to open public key file: %w", err)
		}
		return jwt_provider.NewProvider(privKeyFile, pubKeyFile)
	})

	return i
}

func NewTestInjector(t *testing.T) *do.Injector {
	i := NewInjector()

	do.Override(i, func(_ *do.Injector) (*sql.DB, error) {
		// TODO: Implement do.Shutdownable and do.Healthcheckable
		db, err := sql.Open("sqlite", ":memory:")
		if err != nil {
			return nil, err
		}

		// Load schema
		schemaSqlBytes, err := os.ReadFile(helper.RelativePath("../db/schema.sql"))
		require.NoError(t, err)
		_, err = db.Exec(string(schemaSqlBytes))
		require.NoError(t, err)

		// Load seed
		seedSqlBytes, err := os.ReadFile(helper.RelativePath("../db/seed.sql"))
		require.NoError(t, err)
		_, err = db.Exec(string(seedSqlBytes))
		require.NoError(t, err)

		return db, nil
	})

	do.Override(i, jwt_provider.InjectTestProvider(t))

	return i
}
