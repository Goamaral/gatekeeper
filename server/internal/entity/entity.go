package entity

import (
	"time"

	"github.com/google/uuid"
)

type Challenge struct {
	Id            uint      `json:"-" db:"id"`
	WalletAddress string    `json:"walletAddress" db:"wallet_address"`
	Token         string    `json:"token" db:"token"`
	ExpiredAt     time.Time `json:"expiredAt" db:"expired_at"`
}

type Company struct {
	Uuid             uuid.UUID `json:"uuid" db:"uuid"`
	CreatedAt        time.Time `json:"createdAt" db:"created_at"`
	ApiKey           string    `json:"apiKey" db:"api_key"`
	AdminAccountUuid uuid.UUID `json:"adminAccountUuid" db:"admin_account_uuid"`
}

type Account struct {
	Uuid          uuid.UUID `json:"uuid" db:"uuid"`
	CompanyUuid   uuid.UUID `json:"companyUuid" db:"company_uuid"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	WalletAddress string    `json:"walletAddress" db:"wallet_address"`
	Metadata      []byte    `json:"metadata" db:"metadata"`
}
