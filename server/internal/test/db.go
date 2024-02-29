package test

import (
	"database/sql"
	"gatekeeper/internal/entity"
	"gatekeeper/internal/helper"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const CompanyUuid = "018df6cc-ab90-7592-ae2d-a5c3dd9a79f3"
const ApiKey = "018df6ccab907592ae2da5c3dd9a79f3AFF3MAUaKHt9DVuBBi4Jzw"
const AccountUuid = "018df6cf-44a3-7c50-80fc-055b707fc6d4"
const WalletAddress = "0x25a3aaf7a4fF88A8aa53ff63CFE5e8C16ce93756"

func CreateCompany(t *testing.T, db *sql.DB, adminAccountUuid uuid.UUID) entity.Company {
	companyUuid, err := uuid.NewV7()
	require.NoError(t, err)
	apiKey, err := helper.GenerateApiKey(companyUuid)
	require.NoError(t, err)

	_, err = db.Exec(
		"INSERT INTO companies (uuid, api_key, admin_account_uuid) VALUES (?, ?, ?)",
		companyUuid, apiKey, adminAccountUuid,
	)
	require.NoError(t, err)

	return entity.Company{
		Uuid:             companyUuid,
		CreatedAt:        time.Now(),
		ApiKey:           apiKey,
		AdminAccountUuid: adminAccountUuid,
	}
}

func CreateAccount(t *testing.T, db *sql.DB, companyUuid uuid.UUID, metadata []byte) entity.Account {
	accountUuid, err := uuid.NewV7()
	require.NoError(t, err)
	walletAddress, _ := GenerateWalletAddress(t)

	_, err = db.Exec(
		"INSERT INTO accounts (companyUuid, uuid, wallet_address, metadata) VALUES (?, ?, ?)",
		companyUuid, accountUuid, walletAddress, metadata,
	)
	require.NoError(t, err)

	return entity.Account{
		CompanyUuid:   companyUuid,
		Uuid:          accountUuid,
		CreatedAt:     time.Now(),
		WalletAddress: walletAddress,
		Metadata:      metadata,
	}
}
