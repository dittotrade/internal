package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"
)

// GetDBLock obtains lock for lockInterval period and returns two functions:
// holdLock() continues to hold lock and refreshes it ech refreshInterval.
// It returns status of lock
// cancelLock() releases lock and returns status of operation.
var ErrLockInvalidNumberRecords = errors.New("updated invalid number of records")
var ErrLockRefused = errors.New("lock refused")

func GetDBLock(ctx context.Context, db *sql.DB, lockTable string, lockInterval time.Duration) (
	holdLock, releaseLock func() error) {
	refreshInterval := lockInterval * 5 / 8
	const maxRefreshInterval = time.Second * 5
	if refreshInterval > maxRefreshInterval {
		refreshInterval = maxRefreshInterval
	}
	var lastLock time.Time
	token := strconv.Itoa(rand.Int())
	updateLock := func(query string) (err error) {
		r, err := db.ExecContext(ctx, query)
		if err != nil {
			return fmt.Errorf("query error: %w: %s", err, query)
		}
		var n int64
		n, err = r.RowsAffected()
		if err != nil {
			return err
		}
		if n == 0 {
			err = fmt.Errorf("can't obtain lock %s via %s: %w", lockTable, query, ErrLockRefused)
		}
		if n > 1 {
			err = fmt.Errorf("affected records %d != 1: %w", n, ErrLockInvalidNumberRecords)
		}
		lastLock = time.Now()
		return err
	}
	lockUntil := fmt.Sprintf(`now() + interval '%d millisecond'`, lockInterval.Milliseconds())
	holdLockStm := fmt.Sprintf(`UPDATE %s set locked_until=%s WHERE locked_until < NOW() OR token=%s`, lockTable, lockUntil, token)
	var lastErr error
	holdLock = func() error {
		if lastErr != nil {
			return lastErr
		}
		if err := ctx.Err(); err != nil {
			lastErr = err
			_ = releaseLock()
			lastErr = fmt.Errorf("can't hold lock: context error: %w", err)
			return lastErr
		}
		if time.Since(lastLock) < refreshInterval {
			return nil
		}
		lastErr = updateLock(holdLockStm)
		return lastErr
	}
	releaseLock = func() error {
		if lastErr != nil {
			return nil
		}
		if holdLock() != nil { // can't release lock, it is not ours
			return nil
		}
		_, lastErr = db.Exec("DROP TABLE IF EXISTS " + lockTable)
		if lastErr != nil {
			lastErr = fmt.Errorf("failed to release lock %s: %s", lockTable, lastErr)
			return lastErr
		}
		return nil
	}
	// make sure table exists, create if not
	_, lastErr = db.ExecContext(ctx, `create table if not exists `+lockTable+`(locked_until TIMESTAMP, token bigint)`)
	if lastErr != nil {
		lastErr = fmt.Errorf("failed to create table %s for lock: %w", lockTable, lastErr)
		return
	}
	// make sure there is exactly one record
	var count int
	row := db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", lockTable))
	if err := row.Scan(&count); err != nil {
		lastErr = fmt.Errorf("failed to check row count for table %s : %w", lockTable, err)
		return
	}
	if count > 1 {
		lastErr = fmt.Errorf("on init number of records %d !=1: %w", count, ErrLockInvalidNumberRecords)
		return
	}
	initLock := fmt.Sprintf(`UPDATE %s set token=%s, locked_until=%s WHERE locked_until<now()`, lockTable, token, lockUntil)
	if count == 0 {
		initLock = fmt.Sprintf("INSERT INTO %s(locked_until, token)  VALUES(%s,%s)", lockTable, lockUntil, token)
	}
	lastErr = updateLock(initLock)
	return
}
