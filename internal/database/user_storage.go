package database

import (
	"github.com/sergeychur/technopark_db/internal/models"
)

func (db *DB) GetForumUsers(forumId string) (models.Users, int) {
	return models.Users{}, 0
}

func (db *DB) CreateUser(user models.User) (models.User, int) {
	return models.User{}, 0
}

func (db *DB) GetUser(userNick string) (models.User, int) {
	return models.User{}, 0
}

func (db *DB) UpdateUser(userNick string, user models.UserUpdate) (models.User, int) {
	return models.User{}, 0
}
