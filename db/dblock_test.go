package db

import (
	"context"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestGetDBLock(t *testing.T) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		t.Skip("DATABASE_URL is not set")
	}
	db, err := sqlx.Open("postgres", dbUrl)
	require.NoError(t, err)
	require.NotNil(t, db)
	ctx, cancel := context.WithCancel(context.TODO())
	lockTable := "some_test_lock"
	// drop table before begin
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS "+lockTable)
	require.NoError(t, err)
	lockDuration := time.Millisecond * 40
	holdLock, releaseLock := GetDBLock(ctx, db.DB, lockTable, lockDuration)
	require.NoError(t, holdLock())
	cancel()
	err = holdLock()
	require.Error(t, err)
	require.True(t, errors.Is(err, context.Canceled))
	require.NoError(t, releaseLock())
	_, err = db.ExecContext(ctx, "DROP TABLE "+lockTable)
	require.Error(t, err, "should have dropped the table")
	ctx, cancel = context.WithCancel(context.TODO())
	defer cancel()
	// check error path: can't create table
	require.NoError(t, db.Close())
	holdLock, releaseLock = GetDBLock(ctx, db.DB, lockTable, lockDuration)
	require.Error(t, holdLock(), "db is closed, should return error")
	require.NoError(t, releaseLock())
	// check error path: table records are many
	db, err = sqlx.Open("postgres", dbUrl)
	require.NoError(t, err)
	require.NotNil(t, db)
	r, err := db.ExecContext(ctx, "INSERT INTO "+lockTable+" values (NOW())")
	require.NoError(t, err)
	n, err := r.RowsAffected()
	require.NoError(t, err)
	require.Equal(t, int64(1), n)
	holdLock, releaseLock = GetDBLock(ctx, db.DB, lockTable, lockDuration)
	err = holdLock()
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrLockInvalidNumberRecords))
	require.NoError(t, releaseLock())
	// check can't lock twice
	_, err = db.ExecContext(ctx, "DROP TABLE IF EXISTS "+lockTable)
	require.NoError(t, err)
	holdLock, releaseLock = GetDBLock(ctx, db.DB, lockTable, lockDuration)
	require.NoError(t, holdLock())
	// try to aquire another lock
	time.Sleep(lockDuration / 2)
	holdLock2, releaseLock2 := GetDBLock(ctx, db.DB, lockTable, lockDuration)
	err = holdLock2()
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrLockRefused))
	require.NoError(t, releaseLock2())
	// now wait until lock expires and try to acquire again
	time.Sleep(lockDuration * 5 / 4)
	holdLock2, releaseLock2 = GetDBLock(ctx, db.DB, lockTable, lockDuration)
	err = holdLock2()
	require.NoError(t, err, "lock expired should let another lock to be acquired")
	require.Error(t, holdLock(), "1st lock should have expired")
	err = releaseLock()
	require.NoError(t, err, "not problem even though it locked by other lock")
	// it shoul leave table there
	_, err = db.ExecContext(ctx, "DROP TABLE "+lockTable)
	require.NoError(t, err, "1st release should have left table as it is hold by other lock")
	require.Error(t, releaseLock2(), "releaseLock2 should fail as we dropped the table")
}
