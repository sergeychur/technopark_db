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
	ifExistsUser, err := db.IsUserExist(forum.User)
	if err != nil {
		log.Println(err.Error())
		return forum, DBError
	}
	ifExistsForum := false
	if ifExistsUser {
		ifExistsForum, err = db.IsForumExist(forum.Slug)
		if err != nil {
			log.Println(err.Error())
			return forum, DBError
		}
	}

	if ifExistsForum {
		return forum, Conflict
	}

	if ifExistsUser {
		_, err := db.db.Exec(CreateForum, forum.Slug, forum.Title, forum.User)

		if err != nil {
			log.Println(err)
			return forum, DBError
		}

		return db.GetForum(forum.Slug)
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