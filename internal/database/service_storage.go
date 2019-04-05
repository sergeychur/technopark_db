package database

import (
	"github.com/sergeychur/technopark_db/internal/models"
)

func (db *DB) ClearDB() error {
	return nil
}

func (db *DB) GetDBInfo() (models.Status, int) {
	return models.Status{}, 0
}
