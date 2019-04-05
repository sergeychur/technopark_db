package database

import (
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
)

func (db *DB) CreateForum(forum models.Forum) (models.Forum, int) {
	log.Println("create forum")
	return models.Forum{}, 0
}

func (db *DB) GetForum(ForumId string) (models.Forum, int) {
	log.Println("get forum")
	return models.Forum{}, 0
}
