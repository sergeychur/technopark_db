package database

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

const (
	OK          = 0
	DBError     = 1
	EmptyResult = 2
	Conflict    = 3
)

type DB struct {
	db           *sql.DB
	user         string
	password     string
	databaseName string
	host         string
	port         string
}

func NewDB(user string, password string, dataBaseName string,
	host string, port string) *DB {
	db := new(DB)
	db.user = user
	db.databaseName = dataBaseName
	db.password = password
	db.host = host
	db.port = port
	return db
}

func (db *DB) Start() error {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s "+
		"password=%s dbname=%s sslmode=disable",
		db.host, db.port, db.user, db.password, db.databaseName)
	dataBase, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return err
	}
	db.db = dataBase
	err = db.db.Ping()
	if err != nil {
		db.Close()
		return err
	}
	return nil
}

func (db *DB) Close() {
	_ = db.db.Close()
}

func (db *DB) StartTransaction() (*sql.Tx, error) {
	return db.db.Begin()
}
