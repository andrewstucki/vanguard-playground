package models

import (
	"context"
	_ "embed"
	"strings"
)

//go:embed schema.sql
var schema string

func EnsureSchema(db DBTX) error {
	return execStatements(db, schema)
}

func execStatements(db DBTX, statements string) error {
	for _, statement := range strings.Split(statements, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := db.ExecContext(context.Background(), statement); err != nil {
			return err
		}
	}
	return nil
}
