package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
	"strconv"
	"time"
)

var (
	GetPost            = "SELECT id, author, created, forum, message, parent, thread, is_edited from posts where id = $1"
	UpdatePost         = "UPDATE posts SET message=CASE WHEN $1=''THEN message ELSE $1 END, " +
		"is_edited=CASE WHEN $1='' OR $1=message THEN is_edited ELSE true END WHERE id=$2"
	InsertPost = "INSERT INTO POSTS (message, forum, thread, author, parent, created) VALUES($1, $2, $3, $4, $5, $6) " +
		"RETURNING id, author, created, forum, message, parent, thread, is_edited"
	GetPostsFlatPart1       = "SELECT id, author, created, forum, message, parent, thread, is_edited FROM posts WHERE thread = $1 "
	GetPostsFlatSincePart = "AND id %s $3 "
	GetPostsFlatPart2 = "ORDER BY id %s LIMIT $2"
	GetPostsTreePart1       = "SELECT id, author, created, forum, message, parent, thread, is_edited FROM posts WHERE thread = $1 "
	GetPostsTreePart2 = 	"ORDER BY path %s LIMIT $2"
	GetPostsTreeSincePart = "AND path %s (SELECT path FROM posts WHERE id = $3) "
	GetPostsParentTreePart1 = "SELECT id, author, created, forum, message, parent, thread, is_edited FROM posts WHERE thread = $1 "
	//GetPostsParentTreePart2 = "ORDER BY path[1] %s, path "
	GetPostsParentTreePart2 = "ORDER BY path "
	GetPostsParentTreePart2Alt = "AND path[1] IN (SELECT id FROM posts WHERE thread=$1 AND parent=0 ORDER BY id %s LIMIT $2) ORDER BY path[1] %s, path"
	GetPostsParentTreeSincePart = "AND path[1] IN (SELECT id FROM posts WHERE parent=0 AND id %s $3 ORDER BY id %s LIMIT $2) "
	// LIMIT $2
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
	subqueries := map[string]bool{
		"user":   false,
		"forum":  false,
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
	tx, err := db.StartTransaction()
	if err != nil {
		return models.Posts{}, DBError
	}
	defer tx.Rollback()
	forumId, threadId, retVal := GetThreadForumBySlug(tx, slug)
	if retVal != OK {
		return nil, retVal
	}
	postsToReturn := make(models.Posts, 0)
	currentTime := time.Now()
	timeString := currentTime.Format(time.RFC3339)
	for _, post := range posts {
		ifUserExist, err := IsUserExist(tx, post.Author)
		if post.Parent != 0 {
			ifParentExist := false
			err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM POSTS WHERE id = $1 AND thread = $2)", post.Parent, threadId).Scan(&ifParentExist)
			if err != nil {
				return nil, DBError
			}
			if !ifParentExist {
				return nil, Conflict
			}
		}

		if !ifUserExist {
			return nil, EmptyResult
		}
		curPost := models.Post{}
		err = tx.QueryRow(InsertPost, post.Message, forumId, threadId, post.Author, post.Parent,
			timeString).Scan(&curPost.ID, &curPost.Author, &curPost.Created,
			&curPost.Forum, &curPost.Message, &curPost.Parent, &curPost.Thread, &curPost.IsEdited)
		if err != nil {
			return nil, DBError
		}
		postsToReturn = append(postsToReturn, &curPost)
	}
	err = tx.Commit()
	if err != nil {
		return nil, DBError
	}
	return postsToReturn, OK
}

func (db *DB) CreatePostsById(id string, posts models.Posts) (models.Posts, int) {
	log.Println("create posts by slug")
	tx, err := db.StartTransaction()
	if err != nil {
		return models.Posts{}, DBError
	}
	defer tx.Rollback()
	forumId, retVal := GetThreadForumById(tx, id)
	if retVal != OK {
		return nil, retVal
	}
	postsToReturn := make(models.Posts, 0)
	currentTime := time.Now()
	timeString := currentTime.Format(time.RFC3339)
	for _, post := range posts {
		ifUserExist, err := IsUserExist(tx, post.Author)
		if err != nil {
			return nil, DBError
		}
		if post.Parent != 0 {
			ifParentExist := false
			err := tx.QueryRow("SELECT EXISTS(SELECT 1 FROM POSTS WHERE id = $1 AND thread = $2)", post.Parent, id).Scan(&ifParentExist)
			if err != nil {
				return nil, DBError
			}
			if !ifParentExist {
				return nil, Conflict
			}
		}
		if !ifUserExist {
			return nil, EmptyResult
		}
		curPost := models.Post{}
		err = tx.QueryRow(InsertPost, post.Message, forumId, id, post.Author, post.Parent,
			timeString).Scan(&curPost.ID, &curPost.Author, &curPost.Created,
			&curPost.Forum, &curPost.Message, &curPost.Parent, &curPost.Thread, &curPost.IsEdited)
		if err != nil {
			return nil, DBError
		}
		postsToReturn = append(postsToReturn, &curPost)
	}
	err = tx.Commit()
	if err != nil {
		return nil, DBError
	}
	return postsToReturn, OK
}

func (db *DB) GetPostsBySlug(slug string, limit string, since string,
	sort string, desc string) (models.Posts, int) {
	log.Println("get posts by slug")
	log.Printf("Params: %s, %s, %s,%s, %s", slug, limit, since, sort, desc)
	id := ""
	row := db.db.QueryRow(getThreadIdBySlug, slug)
	err := row.Scan(&id)
	if err == sql.ErrNoRows {
		return nil, EmptyResult
	}
	if err != nil {
		return nil, DBError
	}
	switch sort {
	case "flat":
		return db.GetPostsFlat(id, limit, since, desc)
	case "tree":
		return db.GetPostsTree(id, limit, since, desc)
	case "parent_tree":
		return db.GetPostsParentTree(id, limit, since, desc)
	}
	return models.Posts{}, OK
}

func (db *DB) GetPostsById(id string, limit string, since string,
	sort string, desc string) (models.Posts, int) {
	log.Println("get posts by slug")
	log.Printf("Params: %s, %s, %s,%s, %s", id, limit, since, sort, desc)

	switch sort {
	case "flat":
		return db.GetPostsFlat(id, limit, since, desc)
	case "tree":
		return db.GetPostsTree(id, limit, since, desc)
	case "parent_tree":
		return db.GetPostsParentTree(id, limit, since, desc)
	}
	return models.Posts{}, OK
}

func (db *DB) GetPostsFlat(id string, limit string, since string, desc string) (models.Posts, int) {
	ifDesc, _ := strconv.ParseBool(desc) // mb check error
	strDesc := "ASC"
	if ifDesc {
		strDesc = "DESC"
	}
	rows := &sql.Rows{}
	err := errors.New("")
	if since != "" {
		actualSince := ""
		if ifDesc {
			actualSince = fmt.Sprintf(GetPostsFlatSincePart, "<")
		} else {
			actualSince = fmt.Sprintf(GetPostsFlatSincePart, ">")
		}
		query := GetPostsFlatPart1 + actualSince + GetPostsFlatPart2
		rows, err = db.db.Query(fmt.Sprintf(query, strDesc), id, limit, since)
	} else {
		query := GetPostsFlatPart1 + GetPostsFlatPart2
		rows, err = db.db.Query(fmt.Sprintf(query, strDesc), id, limit)
	}
	if err != nil {
		return nil, DBError
	}
	defer rows.Close()
	posts := make(models.Posts, 0)
	for rows.Next() {
		post := new(models.Post)
		err := rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		posts = append(posts, post)
	}
	return posts, OK
}

func (db *DB) GetPostsTree(id string, limit string, since string, desc string) (models.Posts, int) {
	ifDesc, _ := strconv.ParseBool(desc) // mb check error
	strDesc := "ASC"
	if ifDesc {
		strDesc = "DESC"
	}
	rows := &sql.Rows{}
	err := errors.New("")
	if since != "" {
		actualSince := ""
		if ifDesc {
			actualSince = fmt.Sprintf(GetPostsTreeSincePart, "<")
		} else {
			actualSince = fmt.Sprintf(GetPostsTreeSincePart, ">")
		}
		query := GetPostsTreePart1 + actualSince + GetPostsTreePart2
		rows, err = db.db.Query(fmt.Sprintf(query, strDesc), id, limit, since)
	} else {
		query := GetPostsTreePart1 + GetPostsTreePart2
		rows, err = db.db.Query(fmt.Sprintf(query, strDesc), id, limit)
	}
	//rows, err := db.db.Query(fmt.Sprintf(getPostsTree, strDesc), id, limit, since)
	if err != nil {
		return nil, DBError
	}
	defer rows.Close()
	posts := make(models.Posts, 0)
	for rows.Next() {
		post := new(models.Post)
		err := rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		posts = append(posts, post)
	}
	return posts, OK
}

func (db *DB) GetPostsParentTree(id string, limit string, since string, desc string) (models.Posts, int) {
	ifDesc, _ := strconv.ParseBool(desc) // mb check error
	strDesc := "ASC"
	if ifDesc {
		strDesc = "DESC"
	}
	rows := &sql.Rows{}
	err := errors.New("")
	query := ""
	if since != "" {
		actualSince := ""
		if ifDesc {
			actualSince = fmt.Sprintf(GetPostsParentTreeSincePart, "<=", "desc")
		} else {
			actualSince = fmt.Sprintf(GetPostsParentTreeSincePart, ">=", "asc")
		}
		query = GetPostsParentTreePart1 + actualSince + GetPostsParentTreePart2
		rows, err = db.db.Query(query, id, limit, since)
	} else {
		query = GetPostsParentTreePart1 + GetPostsParentTreePart2Alt
		query = fmt.Sprintf(query, strDesc, strDesc)
		rows, err = db.db.Query(query, id, limit)
	}
	if err != nil {
		return nil, DBError
	}
	defer rows.Close()
	posts := make(models.Posts, 0)
	for rows.Next() {
		post := new(models.Post)
		err := rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		posts = append(posts, post)
	}
	return posts, OK
}
