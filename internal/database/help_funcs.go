package database

import (
	"database/sql"
	"fmt"
)

const (
	Check = "SELECT EXISTS( SELECT 1 FROM %s WHERE %s = $1)"
	//CheckForum = "SELECT EXISTS( SELECT 1 FROM forum WHERE slug = $1)"
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

func IsThreadExist(tx *sql.Tx, slug string) (bool, error) {
	return IsExist(tx, slug, "slug", "threads")
}
