package database

import (
	"database/sql"
	"fmt"
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
)

const (
	createThread         = "INSERT INTO threads (slug, title, author, forum, message) VALUES($1, $2, $3, $4, $5)"
	createThreadWithTime = "INSERT INTO threads (slug, created, title, author, forum, message) VALUES($1, $2, $3, $4, $5, $6)"
	getThreadBySlug      = "SELECT * FROM threads WHERE slug = $1"
	getThreadById        = "SELECT * FROM threads WHERE id = $1"
	getForumThreads      = "SELECT * FROM threads WHERE forum = $1 AND created >= $2 ORDER BY created %s LIMIT $3" // mb change for sql from lections
	updateThreadBySlug   = "UPDATE threads SET message=$1, title=$2 WHERE slug=$3 AND message != $1 AND message != ''"
	updateThreadById     = "UPDATE threads SET message=$1, title=$2 WHERE id=$3 AND message != $1 AND message != ''"
	voteThread           = "INSERT INTO votes(thread, author, is_like) VALUES($1, $2, $3) ON CONFLICT (thread, author) " +
		"DO UPDATE SET is_like = $3"
)

const (
	LIKE    = true
	DISLIKE = false
)

func (db *DB) CreateThread(thread models.Thread, forumId string) (models.Thread, int) {
	tx, err := db.StartTransaction()
	defer tx.Rollback()
	if err != nil {
		log.Println(err)
		return models.Thread{}, DBError
	}
	ifExistsUser, err := IsUserExist(tx, thread.Author)
	if err != nil {
		log.Println(err.Error())
		return models.Thread{}, DBError
	}
	ifExistsForum, err := IsForumExist(tx, forumId)
	if !ifExistsUser || !ifExistsForum {
		return models.Thread{}, EmptyResult
	}
	ifExistsThread := false
	if thread.Slug != "" {
		ifExistsThread, err = IsThreadExistBySlug(tx, thread.Slug)
		if err != nil {
			log.Println(err.Error())
			return models.Thread{}, DBError
		}
	}
	if ifExistsThread {
		return models.Thread{}, Conflict
	}

	if thread.Created != "" {
		_, err = tx.Exec(createThreadWithTime, thread.Slug, thread.Created,
			thread.Title, thread.Author, forumId, thread.Message)
	} else {
		_, err = tx.Exec(createThread, thread.Slug, thread.Title, thread.Author, forumId, thread.Message)
	}

	// TODO(Me): Deal with UNIQUE on thread, some shit now
	if err != nil {
		log.Println(err)
		return models.Thread{}, Conflict
	}
	err = tx.Commit()
	if err != nil {
		return models.Thread{}, DBError
	}
	if thread.Slug != "" {
		return db.GetThreadBySlug(thread.Slug)
	}
	return thread, OK
}

func (db *DB) GetForumThreads(forumId string, limit string,
	since string, desc string) (models.Threads, int) {
	log.Println("get forum threads")

	rows, err := db.db.Query(fmt.Sprintf(getForumThreads, desc), forumId, since, limit)
	if err != nil {
		return models.Threads{}, DBError
	}
	defer rows.Close()
	threads := models.Threads{}
	i := 0
	for rows.Next() {
		i++
		thread := new(models.Thread)
		err := rows.Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum,
			&thread.Message, &thread.Slug, &thread.Title, &thread.Votes)
		if err != nil {
			return models.Threads{}, DBError
		}
		threads = append(threads, thread)
	}
	if i == 0 {
		return models.Threads{}, EmptyResult
	}
	return threads, OK
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
	slug := sql.NullString{}
	err := row.Scan(&thread.ID, &thread.Author, &thread.Created, &thread.Forum,
		&thread.Message, &slug, &thread.Title, &thread.Votes)
	if slug.Valid {
		thread.Slug = slug.String
	}
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
	tx, err := db.StartTransaction()
	defer tx.Rollback()
	if err != nil {
		return models.Thread{}, DBError
	}
	ifThreadExist, err := IsThreadExistBySlug(tx, slug)
	if !ifThreadExist {
		return models.Thread{}, EmptyResult
	}
	_, err = tx.Exec(updateThreadBySlug, update.Message, update.Title, slug)
	if err != nil {
		return models.Thread{}, DBError
	}

	err = tx.Commit()
	if err != nil {
		return models.Thread{}, DBError
	}

	return db.GetThreadBySlug(slug)
}

func (db *DB) UpdateThreadById(id string, update models.ThreadUpdate) (models.Thread, int) {
	tx, err := db.StartTransaction()
	defer tx.Rollback()
	if err != nil {
		return models.Thread{}, DBError
	}
	ifThreadExist, err := IsThreadExistById(tx, id)
	if !ifThreadExist {
		return models.Thread{}, EmptyResult
	}
	_, err = tx.Exec(updateThreadById, update.Message, update.Title, id)
	if err != nil {
		return models.Thread{}, DBError
	}

	err = tx.Commit()
	if err != nil {
		return models.Thread{}, DBError
	}

	return db.GetThreadById(id)
}

func (db *DB) VoteBySlug(slug string, vote models.Vote) (models.Thread, int) {
	tx, err := db.StartTransaction()
	defer tx.Rollback()
	if err != nil {
		return models.Thread{}, DBError
	}
	id, stat := GetThreadIdBySlug(tx, slug)
	if stat != OK {
		return models.Thread{}, stat
	}

	ifUserExist, err := IsUserExist(tx, vote.Nickname)
	if !ifUserExist {
		return models.Thread{}, EmptyResult
	}
	voice := LIKE
	if vote.Voice == -1 {
		voice = DISLIKE
	}
	_, err = tx.Exec(voteThread, id, vote.Nickname, voice)
	if err != nil {
		return models.Thread{}, DBError
	}

	err = tx.Commit()
	if err != nil {
		return models.Thread{}, DBError
	}
	return db.GetThreadById(id)
}

func (db *DB) VoteById(id string, vote models.Vote) (models.Thread, int) {
	tx, err := db.StartTransaction()
	defer tx.Rollback()
	if err != nil {
		return models.Thread{}, DBError
	}
	ifThreadExist, err := IsThreadExistById(tx, id)
	if !ifThreadExist {
		return models.Thread{}, EmptyResult
	}
	ifUserExist, err := IsUserExist(tx, vote.Nickname)
	if !ifUserExist {
		return models.Thread{}, EmptyResult
	}
	voice := LIKE
	if vote.Voice == -1 {
		voice = DISLIKE
	}
	_, err = tx.Exec(voteThread, id, vote.Nickname, voice)
	if err != nil {
		return models.Thread{}, DBError
	}

	err = tx.Commit()
	if err != nil {
		return models.Thread{}, DBError
	}
	return db.GetThreadById(id)
}
