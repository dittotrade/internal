package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/dittotrade/internal/utils"
	"github.com/jmoiron/sqlx"
	"log"
	"os"
	"reflect"
	"strings"
)

type (
	Record map[string]interface{}
	Mockup struct {
		Records map[string][]interface{}
		// private mechanics
		ctx context.Context
		dbx *sqlx.DB
	}
)

func (dbm *Mockup) AddRecord(tableName string, defaults, rec Record, onConflict string, retvalues interface{}) (
	err error) {
	// calculate returning fields from retvalues reflectin
	var returning string
	if retvalues != nil {
		// get list of fields
		stru := reflect.Indirect(reflect.ValueOf(retvalues))
		rr := stru.Type()
		var retfields []string
		for i := 0; i < rr.NumField(); i++ {
			retfields = append(retfields, utils.Underscore(rr.Field(i).Name))
		}
		returning = " returning " + strings.Join(retfields, ",")
	}
	var fields []string
	var values []interface{}
	for k, v := range defaults {
		if _, ok := rec[k]; !ok {
			rec[k] = v
		}
	}
	for k, v := range rec {
		values = append(values, v)
		fields = append(fields, k)
	}
	c := len(fields)
	placeHolders := GetPlaceHolders(c)
	if onConflict != "" {
		var setFields []string
		for _, f := range fields {
			setFields = append(setFields, f+"= excluded."+f)
		}
		onConflict = " on conflict(" + onConflict + ") do update set " + strings.Join(setFields, ",")
	}
	// build statement
	query := fmt.Sprintf("INSERT INTO %s(%s) VALUES (%s)",
		tableName, strings.Join(fields, ","), placeHolders) + onConflict + returning
	err = dbm.dbx.GetContext(dbm.ctx, retvalues, query, values...)
	if err != nil {
		return
	}
	dbm.Records[tableName] = append(dbm.Records[tableName], retvalues)
	return
}

func TryLocalDB(ctx context.Context) *Mockup {
	var dbm Mockup
	tryURLs := []string{
		"postgresql://pguser:pgpass@127.0.0.1:5431/pgdb?sslmode=disable",
		os.Getenv("LOCAL_DATABASE_URL"),
		os.Getenv("DATABASE_URL"),
	}
	var err error
	for _, uri := range tryURLs {
		err = dbm.OpenDB(ctx, uri)
		if err == nil {
			return &dbm
		}
	}
	return nil
}

var ErrNoDatabaseURL = errors.New("db url not defined: DATABASE_URL is not set")

func (dbm *Mockup) OpenDB(actx context.Context, dbUrl string) (err error) {
	dbm.ctx = actx
	if dbUrl == "" {
		return ErrNoDatabaseURL
	}
	dbm.dbx, err = sqlx.Open("postgres", dbUrl)
	if err != nil {
		return fmt.Errorf("open DB: %w", err)
	}
	dbm.dbx.MapperFunc(utils.Underscore)
	dbm.Records = make(map[string][]interface{})
	return nil
}

func CleanRecords(dbm *Mockup, tableName string, records []interface{}) {
	var ids []string
	for _, r := range records {
		rr := reflect.ValueOf(r).Elem()
		// id field should go first. this is easier to enforce than Id
		v := rr.Field(0).Interface()
		id := fmt.Sprintf("'%v'", v)
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return
	}
	r, err := dbm.dbx.ExecContext(dbm.ctx, fmt.Sprintf("DELETE FROM %s where id in (%s)", tableName, strings.Join(ids, ",")))
	var ra int64
	if err == nil {
		ra, err = r.RowsAffected()
	}
	if err != nil {
		log.Printf("remove records from %s: %s", tableName, err)
	} else {
		log.Printf("removed %d records from %s", ra, tableName)
	}
}

func (dbm *Mockup) Close() {
	// remove created mockup records
	for tableName, records := range dbm.Records {
		CleanRecords(dbm, tableName, records)
	}
}
