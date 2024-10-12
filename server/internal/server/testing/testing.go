package server_testing

import (
	"crypto/ecdsa"
	"database/sql"
	"gatekeeper/internal/entity"
	"gatekeeper/internal/helper"
	"gatekeeper/pkg/jwt_provider"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/golang-jwt/jwt/v5"
	"github.com/samber/do"
	"github.com/stretchr/testify/require"
)

const ApiKey = "018df6ccab907592ae2da5c3dd9a79f3AFF3MAUaKHt9DVuBBi4Jzw"
const WalletAddress = "0x25a3aaf7a4fF88A8aa53ff63CFE5e8C16ce93756"

func CreateCompany(t *testing.T, i *do.Injector, adminAccountId uint) entity.Company {
	db := do.MustInvoke[*sql.DB](i)

	apiKey, err := helper.GenerateApiKey()
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

func CreateAccount(t *testing.T, i *do.Injector, companyId uint, metadata []byte) entity.Account {
	db := do.MustInvoke[*sql.DB](i)
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

func GenerateWalletAddress(t *testing.T) (string, *ecdsa.PrivateKey) {
	privateKey, err := crypto.GenerateKey()
	require.NoError(t, err)

	publicKey := privateKey.Public()
	address := crypto.PubkeyToAddress(*publicKey.(*ecdsa.PublicKey))

	return address.Hex(), privateKey
}

func GenerateProofToken(t *testing.T, i *do.Injector, walletAddress string, expiredAt time.Time) string {
	jwtProvider := do.MustInvoke[jwt_provider.Provider](i)
	proofToken, err := jwtProvider.GenerateSignedToken(jwt.RegisteredClaims{
		Subject:   walletAddress,
		ExpiresAt: &jwt.NumericDate{Time: expiredAt},
	})
	require.NoError(t, err)
	return proofToken
}
