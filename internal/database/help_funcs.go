package database

import "fmt"

const (
	Check = "SELECT EXISTS( SELECT 1 FROM %s WHERE %s = $1)"
	//CheckForum = "SELECT EXISTS( SELECT 1 FROM forum WHERE slug = $1)"
)

func (db *DB) IsExist(pk string, pkName string, table string) (bool, error) {
	ifExists := false
	row := db.db.QueryRow(fmt.Sprintf(Check, table, pkName), pk)
	err := row.Scan(&ifExists)
	if err != nil {
		return ifExists, err
	}
	return ifExists, nil
}

func (db *DB) IsUserExist(userNick string) (bool, error) {
	return db.IsExist(userNick, "nick_name", "users")
}

func (db *DB) IsForumExist(slug string) (bool, error) {
	return db.IsExist(slug, "slug", "forum")
}

func (db *DB) IsThreadExist(slug string) (bool, error) {
	return db.IsExist(slug, "slug", "threads")
}
