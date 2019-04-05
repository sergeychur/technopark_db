package database

import (
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
)

func (db *DB) GetPostInfo(postId string) (models.PostFull, int) {
	log.Println("get post info")
	return models.PostFull{}, 0
}

func (db *DB) UpdatePost(postId string, update models.PostUpdate) (models.Post, int) {
	log.Println("update post info")
	return models.Post{}, 0
}

func (db *DB) CreatePostsBySlug(slug string, posts models.Posts) (models.Posts, int) {
	log.Println("create posts by slug")
	return models.Posts{}, 0
}

func (db *DB) CreatePostsById(id string, posts models.Posts) (models.Posts, int) {
	return models.Posts{}, 0
}

func (db *DB) GetPostsBySlug(slug string, limit string, since string,
	sort string, desc string) (models.Posts, int) {
	log.Println("get posts by slug")
	log.Printf("Params: %s, %s, %s,%s, %s", slug, limit, since, sort, desc)
	return models.Posts{}, 0
}

func (db *DB) GetPostsById(id string, limit string, since string,
	sort string, desc string) (models.Posts, int) {
	return models.Posts{}, 0
}
