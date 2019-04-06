package database

import (
	"database/sql"
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
)

const (
	GetForum    = "SELECT * FROM forum where slug = $1"
	CreateForum = "INSERT INTO forum (slug, title, user_nick) VALUES($1, $2, $3)"
)

func (db *DB) CreateForum(forum models.Forum) (models.Forum, int) {
	log.Println("create forum")
	tx, err := db.StartTransaction()
	if err != nil {
		return models.Forum{}, DBError
	}
	defer tx.Rollback()
	ifExistsUser, err := IsUserExist(tx, forum.User)
	if err != nil {
		log.Println(err.Error())
		return forum, DBError
	}
	ifExistsForum := false
	if ifExistsUser {
		ifExistsForum, err = IsForumExist(tx, forum.Slug)
		if err != nil {
			log.Println(err.Error())
			return forum, DBError
		}
	}

	if ifExistsForum {
		return forum, Conflict
	}

	if ifExistsUser {
		_, err := tx.Exec(CreateForum, forum.Slug, forum.Title, forum.User)

		if err != nil {
			log.Println(err)
			return forum, DBError
		}
		err = tx.Commit()
		forum, resVal := db.GetForum(forum.Slug)
		if err != nil {
			return forum, DBError
		}
		return forum, resVal
	}
	return models.Forum{}, EmptyResult
}

func (db *DB) GetForum(ForumId string) (models.Forum, int) {
	log.Println("get forum")
	row := db.db.QueryRow(GetForum, ForumId)
	forum := models.Forum{}
	err := row.Scan(&forum.Posts, &forum.Slug, &forum.Threads, &forum.Title, &forum.User)
	if err == sql.ErrNoRows {
		return forum, EmptyResult
	}
	if err != nil {
		log.Println(err.Error())
		return forum, DBError
	}
	return forum, OK
}
