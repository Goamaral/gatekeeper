package internal

import (
	"gatekeeper/pkg/db"
	"os"
	"testing"

	"github.com/samber/do"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	_ "github.com/glebarez/go-sqlite"
)

func NewInjector() *do.Injector {
	injector := do.New()

	do.Provide(injector, func(_ *do.Injector) (db.Provider, error) {
		// TODO: Implement do.Shutdownable and do.Healthcheckable
		return db.NewProvider("sqlite", "./db/database.sqlite")
	})

	do.Provide(injector, func(_ *do.Injector) (*logrus.Logger, error) {
		logger := logrus.New()
		logger.SetFormatter(&logrus.JSONFormatter{})
		return logger, nil
	})

	return injector
}

func NewTestInjector(t *testing.T) *do.Injector {
	injector := NewInjector()

	do.Override(injector, func(_ *do.Injector) (db.Provider, error) {
		// TODO: Implement do.Shutdownable and do.Healthcheckable
		p, err := db.NewProvider("sqlite", ":memory:")
		if err != nil {
			return db.Provider{}, err
		}

		// Load schema
		schemaSqlBytes, err := os.ReadFile(RelativePath("../db/schema.sql"))
		require.NoError(t, err)

		_, err = p.DB.Exec(string(schemaSqlBytes))
		require.NoError(t, err)

		return p, nil
	})

	return injector
}
