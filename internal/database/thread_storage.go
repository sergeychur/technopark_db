package database

import (
	"github.com/sergeychur/technopark_db/internal/models"
)

func (db *DB) CreateThread(thread models.Thread, forumId string) (models.Thread, int) {
	return models.Thread{}, 0
}

func (db *DB) GetForumThreads(forumId string) (models.Threads, int) {
	return models.Threads{}, 0
}

func (db *DB) GetThreadBySlug(slug string) (models.Thread, int) {
	return models.Thread{}, 0
}

func (db *DB) GetThreadById(id string) (models.Thread, int) {
	return models.Thread{}, 0
}

func (db *DB) UpdateThreadBySlug(slug string, update models.ThreadUpdate) (models.Thread, int) {
	return models.Thread{}, 0
}

func (db *DB) UpdateThreadById(id string, update models.ThreadUpdate) (models.Thread, int) {
	return models.Thread{}, 0
}

func (db *DB) VoteBySlug(slug string, vote models.Vote) (models.Thread, int) {
	return models.Thread{}, 0
}

func (db *DB) VoteById(id string, vote models.Vote) (models.Thread, int) {
	return models.Thread{}, 0
}
