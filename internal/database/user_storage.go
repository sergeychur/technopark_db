package database

import (
	"fmt"
	"github.com/sergeychur/technopark_db/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"log"
)

const (
	getForumUsers = "SELECT u.nick_name, u.about, u.email, u.full_name " +
		"FROM forum_to_users f_u JOIN users u ON (u.nick_name = f_u.user_nick) WHERE f_u.forum = $1 "
	getForumUsersSincePart = "AND u.nick_name %s $2 "
	getForumUsersFinPart   = `ORDER BY u.nick_name  %s LIMIT `
	getUserByNick         = "SELECT * FROM users WHERE nick_name = $1"
	getUsersByEmailOrNick = "SELECT * FROM users WHERE nick_name = $1 OR email = $2"
	createUser            = "INSERT INTO users (nick_name, email, full_name, about) VALUES($1, $2, $3, $4)"
	updateUser            = "UPDATE users SET about=(CASE WHEN $2='' THEN about ELSE $2 END), full_name=(CASE WHEN $3='' THEN full_name ELSE $3 END), " +
		"email=(CASE WHEN $4='' THEN email ELSE $4 END) WHERE nick_name = $1 AND NOT EXISTS (SELECT 1 FROM users WHERE email=$4)"
)

func (db *DB) GetForumUsers(forumId string, limit string,
	since string, desc string) (models.Users, int) {
	query := ""
	rows := &pgx.Rows{}
	ifExist := false
	err := db.db.QueryRow("SELECT TRUE  FROM forum where slug = $1", forumId).Scan(&ifExist)
	if err == pgx.ErrNoRows {
		return nil, EmptyResult
	}
	if err != nil {
		return nil, DBError
	}
	if !ifExist {
		return nil, EmptyResult
	}
	if limit == "" {
		limit = "100"
	}

	if since != "" {
		actualSince := ""
		if desc == "asc" || desc == "" {
			actualSince = fmt.Sprintf(getForumUsersSincePart, ">")
		} else {
			actualSince = fmt.Sprintf(getForumUsersSincePart, "<")
		}
		query = getForumUsers + actualSince + getForumUsersFinPart + "$3"
		rows, err = db.db.Query(fmt.Sprintf(query, desc), forumId, since, limit)
	} else {
		query = getForumUsers + getForumUsersFinPart + "$2"
		rows, err = db.db.Query(fmt.Sprintf(query, desc), forumId, limit)
	}
	if err != nil {
		log.Println(err)
		return models.Users{}, DBError
	}
	defer rows.Close()
	users := models.Users{}
	for rows.Next() {
		user := new(models.User)
		err := rows.Scan(&user.Nickname, &user.About, &user.Email, &user.Fullname)
		if err != nil {
			log.Println(err)
			return models.Users{}, DBError
		}
		users = append(users, user)
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
	rows.Close()
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
	if err == pgx.ErrNoRows {
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
	num := res.RowsAffected()
	if num != 1 {
		return models.User{}, Conflict
	}
	_ = tx.Commit()

	return db.GetUser(userNick)
}
