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
		"ORDER BY nick_name %s LIMIT $3"

	getUserByNick = "SELECT * FROM users WHERE nick_name = $1"
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
		if err != nil{
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

func (db *DB) CreateUser(user models.User) (models.User, int) {
	return models.User{}, 0
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
	return models.User{}, 0
}
