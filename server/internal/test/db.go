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

func CreateCompany(t *testing.T, db *sql.DB, adminAccountId uint) entity.Company {
	companyUuid, err := uuid.NewV7()
	require.NoError(t, err)
	apiKey, err := helper.GenerateApiKey(companyUuid)
	require.NoError(t, err)

	res, err := db.Exec(
		"INSERT INTO companies (api_key, admin_account_id) VALUES (?, ?)",
		apiKey, adminAccountId,
	)
	require.NoError(t, err)

	id, err := res.LastInsertId()
	require.NoError(t, err)

	return entity.Company{
		Id:             uint(id),
		CreatedAt:      time.Now(),
		ApiKey:         apiKey,
		AdminAccountId: adminAccountId,
	}
}

func CreateAccount(t *testing.T, db *sql.DB, companyId uint, metadata []byte) entity.Account {
	walletAddress, _ := GenerateWalletAddress(t)

	_, err := db.Exec(
		"INSERT INTO accounts (company_id, wallet_address, metadata) VALUES (?, ?, ?)",
		companyId, walletAddress, metadata,
	)
	require.NoError(t, err)

	return entity.Account{
		CompanyId:     companyId,
		WalletAddress: walletAddress,
		CreatedAt:     time.Now(),
		Metadata:      metadata,
	}
}
