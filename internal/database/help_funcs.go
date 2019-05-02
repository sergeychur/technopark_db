package database

import (
	"fmt"
	"gopkg.in/jackc/pgx.v2"
	"strconv"
)

const (
	Check                = "SELECT true FROM %s WHERE %s = $1"
	getThreadForumById   = "SELECT forum FROM threads WHERE id = $1"
	getThreadForumBySlug = "SELECT forum, id FROM threads WHERE slug = $1"
	getThreadIdBySlug    = "SELECT id FROM threads WHERE slug = $1"
	getUserNick          = "SELECT nick_name FROM users WHERE nick_name = $1"
	getForumId           = "SELECT slug FROM forum WHERE slug = $1"
)

func IsExist(tx *pgx.Tx, pk string, pkName string, table string) (bool, error) {
	ifExists := false
	row := tx.QueryRow(fmt.Sprintf(Check, table, pkName), pk)
	err := row.Scan(&ifExists)
	if err == pgx.ErrNoRows {
		return ifExists, nil
	}
	if err != nil {
		return ifExists, err
	}
	return ifExists, nil
}

func IsUserExist(tx *pgx.Tx, userNick string) (bool, error) {
	return IsExist(tx, userNick, "nick_name", "users")
}

func IsForumExist(tx *pgx.Tx, slug string) (bool, error) {
	return IsExist(tx, slug, "slug", "forum")
}

func IsThreadExistBySlug(tx *pgx.Tx, slug string) (bool, error) {
	return IsExist(tx, slug, "slug", "threads")
}

func IsThreadExistById(tx *pgx.Tx, id string) (bool, error) {
	return IsExist(tx, id, "id", "threads")
}

func IsPostExist(tx *pgx.Tx, id string) (bool, error) {
	return IsExist(tx, id, "id", "posts")
}

func GetThreadForumBySlug(tx *pgx.Tx, slug string) (string, int, int) {
	ForumId := ""
	threadId := 0
	row := tx.QueryRow(getThreadForumBySlug, slug)
	err := row.Scan(&ForumId, &threadId)
	if err == pgx.ErrNoRows {
		return "", 0, EmptyResult
	}
	if err != nil {
		return ForumId, 0, DBError
	}
	return ForumId, threadId, OK
}

func GetThreadForumById(tx *pgx.Tx, id string) (string, int) {
	ForumId := ""
	row := tx.QueryRow(getThreadForumById, id)
	err := row.Scan(&ForumId)
	if err == pgx.ErrNoRows {
		return "", EmptyResult
	}
	if err != nil {
		return ForumId, DBError
	}
	return ForumId, OK
}

func GetThreadIdBySlug(tx *pgx.Tx, slug string) (string, int) {
	id := 0
	row := tx.QueryRow(getThreadIdBySlug, slug)
	err := row.Scan(&id)
	if err == pgx.ErrNoRows {
		return "", EmptyResult
	}
	if err != nil {
		return "", DBError
	}
	return strconv.Itoa(id), OK
}

func GetUserNick(tx *pgx.Tx, nick string) (string, int) {
	nickName := ""
	row := tx.QueryRow(getUserNick, nick)
	err := row.Scan(&nickName)
	if err == pgx.ErrNoRows {
		return "", EmptyResult
	}
	if err != nil {
		return nickName, DBError
	}
	return nickName, OK
}

func GetForumId(tx *pgx.Tx, forumId string) (string, int) {
	retForumId := ""
	row := tx.QueryRow(getForumId, forumId)
	err := row.Scan(&retForumId)
	if err == pgx.ErrNoRows {
		return "", EmptyResult
	}
	if err != nil {
		return retForumId, DBError
	}
	return retForumId, OK
}
