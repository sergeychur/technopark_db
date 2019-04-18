package database

import (
	"database/sql"
	"fmt"
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
)

var (
	GetPost = "SELECT * from posts where id = $1"
	UpdatePost = "UPDATE posts SET message=$1, is_edited='true' WHERE id=$2 AND message != $1 AND message != ''"
)

func (db *DB) GetPost(postId string) (models.Post, int) {
	post := models.Post{}
	row := db.db.QueryRow(GetPost, postId)
	err := row.Scan(&post.ID, &post.Author, &post.Created,
		&post.Forum, &post.Message, &post.Parent,
		&post.Thread, &post.IsEdited)
	if err == sql.ErrNoRows {
		return post, EmptyResult
	}
	if err != nil {
		return post, DBError
	}
	return post, OK
}

func (db *DB) GetPostInfo(postId string, related []string) (models.PostFull, int) {
	log.Println("get post info")
	subqueries := map[string]bool {
		"user": false,
		"forum": false,
		"thread": false,
	}
	for _, it := range related {
		_, ok := subqueries[it]
		if !ok {
			return models.PostFull{}, DBError
		}
		subqueries[it] = true
	}
	post := models.PostFull{}
	post.Post = new(models.Post)
	reducedPost, retVal := db.GetPost(postId)
	if retVal != OK {
		return post, retVal
	}
	post.Post = &reducedPost

	if subqueries["forum"] {
		post.Forum = new(models.Forum)
		forum, retVal := db.GetForum(post.Post.Forum)
		if retVal != OK {
			return post, DBError
		}
		post.Forum = &forum
	}

	if subqueries["user"] {
		post.Author = new(models.User)
		user, retVal := db.GetUser(post.Post.Author)
		if retVal != OK {
			return post, DBError
		}
		post.Author = &user
	}

	if subqueries["thread"] {
		strId := fmt.Sprintf("%d", post.Post.Thread)
		thread, retVal := db.GetThreadById(strId)
		if retVal != OK {
			return post, DBError
		}
		post.Thread = &thread
	}
	return post, OK
}

func (db *DB) UpdatePost(postId string, update models.PostUpdate) (models.Post, int) {
	log.Println("update post info")
	tx, err := db.StartTransaction()
	defer tx.Rollback()
	if err != nil {
		return models.Post{}, DBError
	}
	ifPostExist, err := IsPostExist(tx, postId)
	if !ifPostExist {
		return models.Post{}, EmptyResult
	}
	_, err = tx.Exec(UpdatePost, update.Message, postId)
	if err != nil {
		return models.Post{}, DBError
	}

	err = tx.Commit()
	if err != nil {
		return models.Post{}, DBError
	}

	return db.GetPost(postId)
}

func (db *DB) CreatePostsBySlug(slug string, posts models.Posts) (models.Posts, int) {
	log.Println("create posts by slug")
	return models.Posts{}, OK
}

func (db *DB) CreatePostsById(id string, posts models.Posts) (models.Posts, int) {
	return models.Posts{}, OK
}

func (db *DB) GetPostsBySlug(slug string, limit string, since string,
	sort string, desc string) (models.Posts, int) {
	log.Println("get posts by slug")
	log.Printf("Params: %s, %s, %s,%s, %s", slug, limit, since, sort, desc)
	return models.Posts{}, OK
}

func (db *DB) GetPostsById(id string, limit string, since string,
	sort string, desc string) (models.Posts, int) {
	return models.Posts{}, OK
}


