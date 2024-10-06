package entity

import (
	"time"
)

type Challenge struct {
	Id            uint      `db:"id"`
	WalletAddress string    `db:"wallet_address"`
	Token         string    `db:"token"`
	ExpiredAt     time.Time `db:"expired_at"`
}

type Company struct {
	Id             uint      `db:"id"`
	CreatedAt      time.Time `db:"created_at"`
	ApiKey         string    `db:"api_key"`
	AdminAccountId uint      `db:"admin_account_id"`
}

type Account struct {
	CompanyId     uint      `db:"company_id"`
	WalletAddress string    `db:"wallet_address"`
	CreatedAt     time.Time `db:"created_at"`
	Metadata      []byte    `db:"metadata"`
}
