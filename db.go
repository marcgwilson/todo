package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"reflect"
)

func OpenDB(name string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	db, err = sql.Open("sqlite3", name)
	if err != nil {
		return nil, err
	}
	if _, err = db.Exec(CreateStmt); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func getFields(i interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(i))
	q := make(map[string]interface{})

	for i := 0; i < v.NumField(); i++ {
		valueField := v.Field(i)
		typeField := v.Type().Field(i)
		if valueField.Kind() == reflect.Ptr && !valueField.IsNil() {
			if tag, ok := typeField.Tag.Lookup("db"); ok {
				if tag != "" && tag != "id" {
					q[tag] = reflect.Indirect(valueField).Interface()
				}
			}
		}
	}

	return q
}
