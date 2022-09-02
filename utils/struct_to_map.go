package utils

import (
	"bytes"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"strings"
)

// StructToURLValues converts any structure into url.Values
func StructToURLValues(v interface{}) url.Values {
	values := url.Values{}
	el := reflect.ValueOf(v)
	if el.Kind() == reflect.Ptr {
		el = el.Elem()
	}
	iVal := el
	ft := iVal.Type()
	for i := 0; i < iVal.NumField(); i++ {
		ft := ft.Field(i)
		name := ft.Tag.Get("json")
		if name == "" {
			name = Underscore(ft.Name)
		} else {
			vals := strings.Split(name, ",")
			if len(vals) > 0 && vals[0] != "" {
				name = vals[0]
			}
		}
		v := fmt.Sprint(iVal.Field(i))
		if len(v) > 0 {
			values.Set(name, v)
		}
	}
	return values
}

// Underscore converts camelCase to underscore (suitable for json & url parameters).
// example Underscore("CamelCase") == "camel_case"
// source https://github.com/buxizhizhoum/inflection/blob/master/inflection.go
var caps2u = regexp.MustCompile(`([A-Z]+)([A-Z][a-z])`)
var lower2u = regexp.MustCompile(`([a-z])([A-Z])`)
var digitL2u = regexp.MustCompile(`(\d)([A-Za-z])`)

func Underscore(s string) string {
	b := []byte(s)
	b = caps2u.ReplaceAll(b, []byte("${1}_${2}"))
	b = lower2u.ReplaceAll(b, []byte("${1}_${2}"))
	b = digitL2u.ReplaceAll(b, []byte("${1}_${2}"))
	b = bytes.ToLower(b)
	return string(b)
}
