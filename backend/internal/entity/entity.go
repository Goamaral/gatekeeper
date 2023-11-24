package entity

import "time"

type Challenge struct {
	Id            uint      `json:"-" db:"id"`
	WalletAddress string    `json:"walletAddress" db:"wallet_address"`
	Token         string    `json:"token" db:"token"`
	ExpiredAt     time.Time `json:"expiredAt" db:"expired_at"`
}
