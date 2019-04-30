package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
	"strconv"
)

const (
	createThread         = "INSERT INTO threads (slug, title, author, forum, message) VALUES($1, $2, $3, $4, $5) RETURNING id"
	createThreadWithTime = "INSERT INTO threads (slug, created, title, author, forum, message) VALUES($1, $2, $3, $4, $5, $6) RETURNING id"
	getThreadBySlug      = "SELECT * FROM threads WHERE slug = $1"
	getThreadById        = "SELECT * FROM threads WHERE id = $1"
	getForumThreadsPart1      = "SELECT * FROM threads WHERE forum = $1 " // mb change for sql from lections
	sincePart = "AND created %s $2 "
	getForumThreadsPart2      = "ORDER BY created %s LIMIT " // mb change for sql from lections
	updateThreadBySlug   = "UPDATE threads SET message=CASE $1 WHEN '' THEN message ELSE $1 END, title=CASE $2 WHEN '' THEN title ELSE $2 END WHERE slug=$3"
	updateThreadById     = "UPDATE threads SET message=CASE $1 WHEN '' THEN message ELSE $1 END, title=CASE $2 WHEN '' THEN title ELSE $2 END WHERE id=$3"
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
	//ifExistsForum, err := IsForumExist(tx, forumId)
	forumId, stat := GetForumId(tx, forumId)
	if stat == DBError {
		return models.Thread{}, stat
	}
	if !ifExistsUser || stat == EmptyResult {
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
		_ = tx.Rollback()
		threadToReturn, stat := db.GetThreadBySlug(thread.Slug)
		if stat != OK {
			return models.Thread{}, stat
		}
		return threadToReturn, Conflict
	}
	insertedId := -1
	if thread.Created != "" {
		row := tx.QueryRow(createThreadWithTime, thread.Slug, thread.Created,
			thread.Title, thread.Author, forumId, thread.Message)
		err := row.Scan(&insertedId)
		if err != nil{
			return models.Thread{}, DBError
		}
	} else {
		row := tx.QueryRow(createThread, thread.Slug, thread.Title, thread.Author, forumId, thread.Message)
		err := row.Scan(&insertedId)
		if err != nil{
			return models.Thread{}, DBError
		}
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
	if insertedId != -1 {
		ID := strconv.Itoa(insertedId)
		return db.GetThreadById(ID)
	}
	return thread, OK
}

func (db *DB) GetForumThreads(forumId string, limit string,
	since string, desc string) (models.Threads, int) {
	log.Println("get forum threads")
	ifExist := false
	err := db.db.QueryRow("SELECT EXISTS(SELECT 1 FROM forum where slug = $1)", forumId).Scan(&ifExist)
	if err != nil {
		return nil, DBError
	}
	if !ifExist {
		return nil, EmptyResult
	}
	query := ""
	rows := &sql.Rows{}
	err = errors.New("")
	actualSince := ""
	if since != "" {
		if desc == "asc" {
			actualSince = fmt.Sprintf(sincePart, ">=")
		} else {
			actualSince = fmt.Sprintf(sincePart, "<=")
		}
		query = getForumThreadsPart1 + actualSince + getForumThreadsPart2 + "$3"
		rows, err = db.db.Query(fmt.Sprintf(query, desc), forumId, since, limit)
	} else {
		query = getForumThreadsPart1 + getForumThreadsPart2 + "$2"
		rows, err = db.db.Query(fmt.Sprintf(query, desc), forumId, limit)
	}
	if err != nil {
		return nil, DBError
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
		return models.Threads{}, OK
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
