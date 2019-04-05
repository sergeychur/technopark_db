package database

import (
	"database/sql"
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
)

const (
	createThread         = "INSERT INTO threads (slug, title, author, forum, message) VALUES($1, $2, $3, $4, $5)"
	createThreadWithTime = "INSERT INTO threads (slug, created, title, author, forum, message) VALUES($1, $2, $3, $4, $5, $6)"
	getThreadBySlug      = "SELECT * FROM threads WHERE slug = $1"
	getThreadById        = "SELECT * FROM threads WHERE id = $1"
)

func (db *DB) CreateThread(thread models.Thread, forumId string) (models.Thread, int) {
	ifExistsUser, err := db.IsUserExist(thread.Author)
	if err != nil {
		log.Println(err.Error())
		return models.Thread{}, DBError
	}
	ifExistsForum, err := db.IsForumExist(forumId)
	if !ifExistsUser || !ifExistsForum {
		return models.Thread{}, EmptyResult
	}
	ifExistsThread := false
	if thread.Slug != "" {
		ifExistsThread, err = db.IsThreadExist(thread.Slug)
		if err != nil {
			log.Println(err.Error())
			return models.Thread{}, DBError
		}
	}
	if ifExistsThread {
		return models.Thread{}, Conflict
	}

	if thread.Created != "" {
		_, err = db.db.Exec(createThreadWithTime, thread.Slug, thread.Created,
			thread.Title, thread.Author, forumId, thread.Message)
	} else {
		_, err = db.db.Exec(createThread, thread.Slug, thread.Title, thread.Author, forumId, thread.Message)
	}

	// TODO(Me): Deal with UNIQUE on thread, some shit now
	if err != nil {
		log.Println(err)
		return models.Thread{}, Conflict
	}
	if thread.Slug != "" {
		return db.GetThreadBySlug(thread.Slug)
	}
	return thread, OK
}

func (db *DB) GetForumThreads(forumId string, limit string,
	since string, desc string) (models.Threads, int) {
	return models.Threads{}, 0
}

func (db *DB) GetThreadBySlug(slug string) (models.Thread, int) {
	log.Println("get thread slug")
	row := db.db.QueryRow(getThreadBySlug, slug)
	thread := models.Thread{}
	err := row.Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum,
		&thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err == sql.ErrNoRows {
		return thread, EmptyResult
	}
	if err != nil {
		log.Println(err.Error())
		return thread, DBError
	}
	return thread, OK
}

func (db *DB) GetThreadById(id string) (models.Thread, int) {
	log.Println("get thread id")
	row := db.db.QueryRow(getThreadById, id)
	thread := models.Thread{}
	err := row.Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum,
		&thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
	if err == sql.ErrNoRows {
		return thread, EmptyResult
	}
	if err != nil {
		log.Println(err.Error())
		return thread, DBError
	}
	return thread, OK
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