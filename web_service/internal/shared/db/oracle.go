package db

import (
	"log"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/sijms/go-ora/v2"
)

func NewOracleDB(driver, dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	// pool config
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	db.MapperFunc(strings.ToLower)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	log.Println("Oracle DB connected")
	return db, nil
}
