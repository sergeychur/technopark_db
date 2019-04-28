package database

import (
	"database/sql"
	"fmt"
)

const (
	Check                = "SELECT EXISTS( SELECT 1 FROM %s WHERE %s = $1)"
	getThreadForumById   = "SELECT forum FROM threads WHERE id = $1"
	getThreadForumBySlug = "SELECT forum, id FROM threads WHERE slug = $1"
	getThreadIdBySlug    = "SELECT id FROM threads WHERE slug = $1"
)

func IsExist(tx *sql.Tx, pk string, pkName string, table string) (bool, error) {
	ifExists := false
	row := tx.QueryRow(fmt.Sprintf(Check, table, pkName), pk)
	err := row.Scan(&ifExists)
	if err != nil {
		return ifExists, err
	}
	return ifExists, nil
}

func IsUserExist(tx *sql.Tx, userNick string) (bool, error) {
	return IsExist(tx, userNick, "nick_name", "users")
}

func IsForumExist(tx *sql.Tx, slug string) (bool, error) {
	return IsExist(tx, slug, "slug", "forum")
}

func IsThreadExistBySlug(tx *sql.Tx, slug string) (bool, error) {
	return IsExist(tx, slug, "slug", "threads")
}

func IsThreadExistById(tx *sql.Tx, id string) (bool, error) {
	return IsExist(tx, id, "id", "threads")
}

func IsPostExist(tx *sql.Tx, id string) (bool, error) {
	return IsExist(tx, id, "id", "posts")
}

func GetThreadForumBySlug(tx *sql.Tx, slug string) (string, int, int) {
	ForumId := ""
	threadId := 0
	row := tx.QueryRow(getThreadForumBySlug, slug)
	err := row.Scan(&ForumId, &threadId)
	if err == sql.ErrNoRows {
		return "", 0, EmptyResult
	}
	if err != nil {
		return ForumId, 0, DBError
	}
	return ForumId, threadId, OK
}

func GetThreadForumById(tx *sql.Tx, id string) (string, int) {
	ForumId := ""
	row := tx.QueryRow(getThreadForumById, id)
	err := row.Scan(&ForumId)
	if err == sql.ErrNoRows {
		return "", EmptyResult
	}
	if err != nil {
		return ForumId, DBError
	}
	return ForumId, OK
}

func GetThreadIdBySlug(tx *sql.Tx, slug string) (string, int) {
	id := ""
	row := tx.QueryRow(getThreadIdBySlug, slug)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return "", EmptyResult
	}
	if err != nil {
		return id, DBError
	}
	return id, OK
}
