package sqlite_ext

import (
	"github.com/glebarez/go-sqlite"
)

func HasErrCode(err error, errCode int) bool {
	sqliteErr, ok := err.(*sqlite.Error)
	if !ok {
		return false
	}
	return sqliteErr.Code() == errCode
}
