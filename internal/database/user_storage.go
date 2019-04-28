package database

import (
	"database/sql"
	"fmt"
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
)

const (
	getForumUsers = "SELECT u.nick_name, u.about, u.email, u.full_name " +
		"FROM users u JOIN posts p ON u.nick_name = p.author " +
		"WHERE p.forum = $1 AND u.nick_name >= $2" +
		"UNION " +
		"SELECT u.nick_name, u.about, u.email, u.full_name " +
		"FROM users u JOIN threads t ON u.nick_name = t.author " +
		"WHERE t.forum = $1 AND u.nick_name >= $2" +
		"ORDER BY nick_name %s LIMIT $3" // change on correct sql from lections

	getUserByNick         = "SELECT * FROM users WHERE nick_name = $1"
	getUsersByEmailOrNick = "SELECT * FROM users WHERE nick_name = $1 OR email = $2"
	createUser            = "INSERT INTO users (nick_name, email, full_name, about) VALUES($1, $2, $3, $4)"
	updateUser            = "UPDATE users SET about=$2, full_name=$3, email=$4 WHERE nick_name = $1 AND NOT EXISTS (SELECT 1 FROM users WHERE email=$4)"
)

func (db *DB) GetForumUsers(forumId string, limit string,
	since string, desc string) (models.Users, int) {
	log.Println("get forum users")

	rows, err := db.db.Query(fmt.Sprintf(getForumUsers, desc), forumId, since, limit)
	if err != nil {
		log.Println(err)
		return models.Users{}, DBError
	}
	defer rows.Close()
	users := models.Users{}
	i := 0
	for rows.Next() {
		i++
		user := new(models.User)
		err := rows.Scan(&user.Nickname, &user.About, &user.Email, &user.Fullname)
		if err != nil {
			log.Println(err)
			return models.Users{}, DBError
		}
		users = append(users, user)
	}
	if i == 0 {
		return models.Users{}, EmptyResult
	}
	return users, OK
}

func (db *DB) CreateUser(user models.User) (models.Users, int) {
	tx, err := db.StartTransaction()
	defer tx.Rollback()
	if err != nil {
		log.Println(err)
		return models.Users{}, DBError
	}
	rows, err := tx.Query(getUsersByEmailOrNick, user.Nickname, user.Email)
	if err != nil {
		log.Println(err)
		return models.Users{}, DBError
	}
	users := make(models.Users, 0)
	i := 0
	for rows.Next() {
		i++
		user := new(models.User)
		err := rows.Scan(&user.Nickname, &user.About, &user.Email, &user.Fullname)
		if err != nil {
			log.Println(err)
			return models.Users{}, DBError
		}
		users = append(users, user)
	}
	_ = rows.Close()
	if i != 0 {
		return users, Conflict
	}
	_, err = tx.Exec(createUser, user.Nickname, user.Email, user.Fullname, user.About)
	if err != nil {
		return nil, DBError
	}
	_ = tx.Commit()
	userToReturn, stat := db.GetUser(user.Nickname)
	users = append(users, &userToReturn)
	return users, stat
}

func (db *DB) GetUser(userNick string) (models.User, int) {
	user := models.User{}
	row := db.db.QueryRow(getUserByNick, userNick)
	err := row.Scan(&user.Nickname, &user.About, &user.Email,
		&user.Fullname)
	if err == sql.ErrNoRows {
		return user, EmptyResult
	}
	if err != nil {
		return user, DBError
	}
	return user, OK
}

func (db *DB) UpdateUser(userNick string, user models.UserUpdate) (models.User, int) {
	tx, err := db.StartTransaction()
	if err != nil {
		return models.User{}, DBError
	}
	defer tx.Rollback()
	ifUserExists, err := IsUserExist(tx, userNick)
	if err != nil {
		return models.User{}, DBError
	}
	if !ifUserExists {
		return models.User{}, EmptyResult
	}
	res, err := tx.Exec(updateUser, userNick, user.About, user.Fullname, user.Email)
	if err != nil {
		return models.User{}, DBError
	}
	num, err := res.RowsAffected()
	if num != 1 {
		return models.User{}, Conflict
	}
	_ = tx.Commit()

	return db.GetUser(userNick)
}
