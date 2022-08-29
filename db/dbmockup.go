package db

import (
	"fmt"
)

type (
	ExchangeAccount struct {
		Name  string
		Eid   int32
		Login int64
		User  *User
	}
	User struct {
		Name  string
		Email string
	}
	TradingAccount struct {
		Name           string
		EchangeAccount *ExchangeAccount
		StopLoss       float64
		Equity         float64
	}
	Mockup struct {
		ExchangeAccounts []*ExchangeAccount
		Users            []*User
		TradingAccounts  []*TradingAccount
		// private mechanics
		cc                  int
		transactionUnderway bool
	}
)

func (dbm *Mockup) NewTradingAccount() *TradingAccount {
	if !dbm.transactionUnderway {
		dbm.cc++
		dbm.transactionUnderway = true
		defer func() { dbm.transactionUnderway = false }()
	}
	var tradingAccount TradingAccount
	tradingAccount.Name = fmt.Sprintf("investment %02d", dbm.cc)
	tradingAccount.EchangeAccount = dbm.NewExchangeAccount()
	dbm.TradingAccounts = append(dbm.TradingAccounts, &tradingAccount)
	return &tradingAccount
}

func (dbm *Mockup) NewUser() *User {
	if !dbm.transactionUnderway {
		dbm.cc++
		dbm.transactionUnderway = true
		defer func() { dbm.transactionUnderway = false }()
	}
	var user User
	user.Name = fmt.Sprintf("user %02d", dbm.cc)
	user.Email = fmt.Sprintf("user%02d@ditto.trade", dbm.cc)
	dbm.Users = append(dbm.Users, &user)
	return &user
}

func (dbm *Mockup) NewExchangeAccount() *ExchangeAccount {
	if !dbm.transactionUnderway {
		dbm.cc++
		dbm.transactionUnderway = true
		defer func() { dbm.transactionUnderway = false }()
	}
	var ea ExchangeAccount
	ea.Name = fmt.Sprintf("a %02d", dbm.cc)
	ea.Eid = int32(dbm.cc)
	ea.Login = int64(dbm.cc)
	ea.User = dbm.NewUser()
	dbm.ExchangeAccounts = append(dbm.ExchangeAccounts, &ea)
	return &ea
}
