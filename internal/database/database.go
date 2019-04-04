package database

import (
	"database/sql"
)

type DB struct {
	db sql.DB
}

func NewDB() *DB {
	db := new(DB)
	return db
}

func (db *DB) Start () {

}

func (db *DB) Close () {

}

