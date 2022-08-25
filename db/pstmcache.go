package db

import (
	"github.com/jmoiron/sqlx"
	"strconv"
	"strings"
	"sync"
)

// utilites to implement caching of prepared statements
func GetPlaceHolders(countArgs int) string {
	placeholders := make([]string, countArgs)
	for i := range placeholders {
		// postgres placeholders are like $1,$2,$3
		placeholders[i] = "$" + strconv.Itoa(i+1)
	}
	return strings.Join(placeholders, ",")
}

// maps fnName to prepared statement, reduces time to compile
var (
	preparedStmt = make(map[string]*sqlx.Stmt)
	psMu         sync.RWMutex
)

// CachePrepared avoid compilation of the same sql statement several times
func CachePrepared(db *sqlx.DB, query string) (stm *sqlx.Stmt, err error) {
	psMu.RLock()
	stm, ok := preparedStmt[query]
	psMu.RUnlock()
	if ok {
		return
	}
	// it takes some time to prepare statement
	// don't block preparedStmt while we are doing this
	stm, err = db.Preparex(query)
	if err != nil {
		return
	}
	// now it is time to write
	psMu.Lock()
	defer psMu.Unlock()
	// still check may be someone else already prepared this statement
	// while we were doing it on our own
	if stmCopy, ok := preparedStmt[query]; ok {
		_ = stm.Close() // close what we prepared
		return stmCopy, nil
	}
	preparedStmt[query] = stm
	return
}

// CloseCashePrepared closes all prepared sql statements cache for cleaning memory at the server
// it is called upong program terminating as well as once in a 5 minute
func CloseCashePrepared() {
	psMu.Lock()
	for k, v := range preparedStmt {
		if v.Close() == nil {
			delete(preparedStmt, k)
		}
	}
	psMu.Unlock()
}
