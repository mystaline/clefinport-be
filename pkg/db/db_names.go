package db

type DBName string

const (
	UserServiceDBName   DBName = "clefinport_user"
	WalletServiceDBName DBName = "clefinport_wallet"
	LogServiceDBName    DBName = "clefinport_log"
)
