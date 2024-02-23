package internal

import (
	"database/sql"
	"fmt"
	"gatekeeper/pkg/fs"
	"gatekeeper/pkg/jwt_provider"
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
		return sql.Open("sqlite", RelativePath("../db/database.sqlite"))
	})

	do.Provide(injector, func(i *do.Injector) (jwt_provider.Provider, error) {
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

	do.Override(injector, jwt_provider.InjectTestProvider(t))

	return injector
}
