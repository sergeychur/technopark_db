package database

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"github.com/sergeychur/technopark_db/internal/models"
	"log"
	"strconv"
	"time"
)

var (
	GetPost            = "SELECT id, author, created, forum, message, parent, thread, is_edited from posts where id = $1"
	UpdatePost         = "UPDATE posts SET message=CASE WHEN $1=''THEN message ELSE $1 END, " +
		"is_edited=CASE WHEN $1='' OR $1=message THEN is_edited ELSE true END WHERE id=$2"
	GetPostsByIds      = "SELECT id, author, created, forum, message, parent, thread, is_edited FROM POSTS WHERE id = ANY($1)"
	CreatePostInThread = "INSERT INTO posts (message, forum, thread, author, parent, created) " +
		"select message, forum, cast (thread as integer), author, cast (parent as bigint), " +
		"cast (created as timestamp)from (values($1, $2, $3, $4, $5, $6)) " +
		"as t (message, forum, thread, author, parent, created) WHERE EXISTS (SELECT 1 FROM posts where cast(id as text)=$5 AND thread::text=$3) OR $5 = '0' RETURNING id"
	GetPostsFlatPart1       = "SELECT id, author, created, forum, message, parent, thread, is_edited FROM posts WHERE thread = $1 "
	GetPostsSincePart = "AND id %s $3 "
	GetPostsFlatPart2 = "ORDER BY id %s LIMIT $2"
	GetPostsTreePart1       = "SELECT id, author, created, forum, message, parent, thread, is_edited FROM posts WHERE thread = $1 "
		/*"AND id >= $3 "*/
	GetPostsTreePart2 = 	"ORDER BY path %s LIMIT $2"
	GetPostsParentTreePart1 = "SELECT id, author, created, forum, message, parent, thread, is_edited FROM posts WHERE thread = $1 "
		/*"AND id >= $3 " +*/
	GetPostsParentTreePart2 = "ORDER BY path[1] %s, path LIMIT $2"
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
	insertedIds := make([]int, 0)
	stmt, err := tx.Prepare(CreatePostInThread)
	if err != nil {
		return nil, DBError
	}
	defer stmt.Close()
	allFound := true
	currentTime := time.Now()
	timeString := currentTime.Format(time.RFC3339)
	postsToReturn := make(models.Posts, 0)
	i := 0
	for _, post := range posts {
		i++
		lastInserted := 0
		err = stmt.QueryRow(post.Message, forumId, threadId, post.Author, post.Parent, timeString).Scan(&lastInserted)
		if err == sql.ErrNoRows {
			allFound = false
			break
		}
		if err != nil {
			return models.Posts{}, DBError
		}
		/*if lastInserted == 0 {
			allFound = false
			break
		}*/
		insertedIds = append(insertedIds, lastInserted)
	}
	if i == 0 {
		return models.Posts{}, OK
	}
	retVal = OK
	if !allFound {
		return nil, Conflict
	}
	_ = stmt.Close()
	_ = tx.Commit()
	rows, err := db.db.Query(GetPostsByIds, pq.Array(insertedIds))
	if err != nil {
		return models.Posts{}, DBError
	}
	defer rows.Close()
	i = 0
	for rows.Next() {
		i++
		post := new(models.Post)
		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum,
			&post.Message, &post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		postsToReturn = append(postsToReturn, post)
	}
	if i == 0 {
		return models.Posts{}, DBError // because there have to be rows
	}
	return postsToReturn, retVal
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
	insertedIds := make([]int, 0)
	stmt, err := tx.Prepare(CreatePostInThread)
	if err != nil {
		return nil, DBError
	}
	defer stmt.Close()
	allFound := true
	currentTime := time.Now()
	timeString := currentTime.Format(time.RFC3339)
	threadId, err := strconv.Atoi(id)
	postsToReturn := make(models.Posts, 0)
	i := 0
	for _, post := range posts {
		i++
		lastInserted := 0
		err = stmt.QueryRow(post.Message, forumId, threadId, post.Author, post.Parent, timeString).Scan(&lastInserted)
		if err != nil {
			return models.Posts{}, DBError
		}
		if lastInserted == 0 {
			allFound = false
			break
		}
		insertedIds = append(insertedIds, lastInserted)
	}
	if i == 0 {
		return models.Posts{}, OK
	}
	retVal = OK
	if !allFound {
		return nil, Conflict
	}
	_ = stmt.Close()
	_ = tx.Commit()
	rows, err := db.db.Query(GetPostsByIds, pq.Array(insertedIds))
	if err != nil {
		return models.Posts{}, DBError
	}
	defer rows.Close()
	i = 0
	for rows.Next() {
		i++
		post := new(models.Post)
		err = rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum,
			&post.Message, &post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		postsToReturn = append(postsToReturn, post)
	}
	if i == 0 {
		return models.Posts{}, DBError // because there have to be rows
	}
	return postsToReturn, retVal
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
			actualSince = fmt.Sprintf(GetPostsSincePart, "<=")
		} else {
			actualSince = fmt.Sprintf(GetPostsSincePart, "=>")
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
	i := 0
	posts := make(models.Posts, 0)
	for rows.Next() {
		i++
		post := new(models.Post)
		err := rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		posts = append(posts, post)
	}
	if i == 0 {
		return nil, EmptyResult // dunno if need
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
			actualSince = fmt.Sprintf(GetPostsSincePart, "<=")
		} else {
			actualSince = fmt.Sprintf(GetPostsSincePart, "=>")
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
	i := 0
	posts := make(models.Posts, 0)
	for rows.Next() {
		i++
		post := new(models.Post)
		err := rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		posts = append(posts, post)
	}
	if i == 0 {
		return nil, EmptyResult // dunno if need
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
	if since != "" {
		actualSince := ""
		if ifDesc {
			actualSince = fmt.Sprintf(GetPostsSincePart, "<=")
		} else {
			actualSince = fmt.Sprintf(GetPostsSincePart, "=>")
		}
		query := GetPostsParentTreePart1 + actualSince + GetPostsParentTreePart2
		rows, err = db.db.Query(fmt.Sprintf(query, strDesc), id, limit, since)
	} else {
		query := GetPostsParentTreePart1 + GetPostsParentTreePart2
		rows, err = db.db.Query(fmt.Sprintf(query, strDesc), id, limit)
	}
	//rows, err := db.db.Query(fmt.Sprintf(getPostsParentTree, strDesc), id, limit, since)
	if err != nil {
		return nil, DBError
	}
	defer rows.Close()
	i := 0
	posts := make(models.Posts, 0)
	for rows.Next() {
		i++
		post := new(models.Post)
		err := rows.Scan(&post.ID, &post.Author, &post.Created, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		posts = append(posts, post)
	}
	if i == 0 {
		return nil, EmptyResult // dunno if need
	}
	return posts, OK
}
