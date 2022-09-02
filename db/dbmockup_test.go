package db

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"log"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestTryLocalDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	dbm := TryLocalDB(ctx)
	if dbm == nil {
		t.Skip("skip test: no db connection")
	}
	defer dbm.Close()
	log.Println("connection to local db: successful")
}

const UserEmail1 = "user01@ditto.trade"

func TestMockup_NewUser(t *testing.T) {
	ctx := context.TODO()
	//ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	//defer cancel()
	dbm := TryLocalDB(ctx)
	if dbm == nil {
		t.Skip("skip test: no db connection")
	}
	defer dbm.Close()
	var userId struct{ Id uuid.UUID }
	err := dbm.AddRecord("users", nil, Record{"email": UserEmail1}, "email", &userId)
	require.NoError(t, err)
	log.Println("user created!", userId.Id)
}

var eaDefaults = Record{
	"login": 1, "account_name": "exchange_account 1", "account_type": "live",
	"currency": "USD", "leverage": 100, "password": "p1", "investor_password": "ip1",
	"kind": "trader",
}

func TestMockup_NewExchangeAccount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	dbm := TryLocalDB(ctx)
	if dbm == nil {
		t.Skip("skip test: no db connection")
	}
	defer dbm.Close()
	var userId struct{ Id uuid.UUID }
	err := dbm.AddRecord("users", nil, Record{"email": UserEmail1}, "email", &userId)
	require.NoError(t, err)
	var ea struct {
		Id          uuid.UUID
		Eid         int32
		Login       int
		AccountName string
		UserId      uuid.UUID
	}
	err = dbm.AddRecord("exchange_accounts", eaDefaults, Record{"user_id": userId.Id}, "login", &ea)
	require.NoError(t, err)
	require.NotZero(t, ea.Eid)
	require.Equal(t, ea.AccountName, "exchange_account 1")
	require.Equal(t, ea.Login, 1)
	require.Equal(t, userId.Id, ea.UserId)
	ea2 := ea
	err = dbm.AddRecord("exchange_accounts", eaDefaults, Record{"login": 2, "account_name": "2", "user_id": userId.Id}, "login", &ea2)
	require.NoError(t, err)
	require.NotZero(t, ea2.Eid)
	require.NotEqual(t, ea.Eid, ea2.Eid)
	require.Equal(t, ea2.Login, 2)
	require.Equal(t, ea2.AccountName, "2")
	require.Equal(t, userId.Id, ea2.UserId)
}

func TestMockup_NewInvestmentAccount(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second*10)
	defer cancel()
	dbm := TryLocalDB(ctx)
	if dbm == nil {
		t.Skip("skip test: no db connection")
	}
	defer dbm.Close()
	var investor, pro struct{ Id uuid.UUID }
	investorEmail := "investor-" + strconv.Itoa(rand.Int()) + "@ditto.trade"
	err := dbm.AddRecord("users", nil, Record{"email": investorEmail}, "email", &investor)
	require.NoError(t, err)
	proTraderEmail := "protrader-" + strconv.Itoa(rand.Int()) + "@ditto.trade"
	err = dbm.AddRecord("users", nil, Record{"email": proTraderEmail}, "email", &pro)
	require.NoError(t, err)
	var investorEA struct {
		Id    uuid.UUID
		Eid   int32
		Login int
	}
	// add investor exchange account
	err = dbm.AddRecord("exchange_accounts", eaDefaults, Record{"user_id": investor.Id}, "login", &investorEA)
	require.NoError(t, err)
	// add pro_trader account
	proEA := investorEA
	err = dbm.AddRecord("exchange_accounts", eaDefaults, Record{"login": 2, "account_name": proTraderEmail, "kind": "pro_trader",
		"user_id": pro.Id},
		"login", &proEA)
	require.NoError(t, err)
	// create investment account
	var ia struct {
		Id       uuid.UUID
		StopLoss float64
	}
	iaData := Record{"user_id": investor.Id, "status": "active", "name": "ia 1", "login": proEA.Login, "eid": proEA.Eid,
		"stop_loss": 90}
	err = dbm.AddRecord("investment_accounts", nil, iaData, "", &ia)
	require.NoError(t, err)
	require.Equal(t, float64(90), ia.StopLoss)
}
