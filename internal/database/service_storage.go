package database

import (
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
)

const (
	TruncateAllTables = "TRUNCATE votes, posts, threads, forum, users"
	GetDBInfo = "SELECT count_forum, count_post, count_thread, count_user FROM " +
		"(SELECT COUNT(*) AS count_forum FROM forum) AS count1, " +
		"(SELECT COUNT(*) AS count_post FROM posts) AS count2, " +
		"(SELECT COUNT(*) AS count_thread FROM threads) AS count3, " +
		"(SELECT COUNT(*) AS count_user FROM users) AS count4"

)

func (db *DB) ClearDB() error {
	tx, err := db.StartTransaction()
	if err != nil {
		log.Println(err)
		return err
	}
	defer tx.Rollback()
	_, err = tx.Exec(TruncateAllTables)

	if err != nil {
		log.Println(err)
		return err
	}
	err = tx.Commit()
	return err
}

func (db *DB) GetDBInfo() (models.Status, int) {
	log.Println("get status")
	row := db.db.QueryRow(GetDBInfo)
	status := models.Status{}
	err := row.Scan(&status.Forum, &status.Post, &status.Thread, &status.User)
	if err != nil {
		log.Println(err.Error())
		return status, DBError
	}
	return status, OK
}
