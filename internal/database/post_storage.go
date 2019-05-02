package database

import (
	"errors"
	"fmt"
	"github.com/sergeychur/technopark_db/internal/models"
	"gopkg.in/jackc/pgx.v2"
	"strconv"
	"time"
)

var (
	GetPost    = "SELECT id, author, created, forum, message, parent, thread, is_edited from posts where id = $1"
	UpdatePost = "UPDATE posts SET message=CASE WHEN $1=''THEN message ELSE $1 END, " +
		"is_edited=CASE WHEN $1='' OR $1=message THEN is_edited ELSE true END WHERE id=$2"
	InsertPost = "INSERT INTO POSTS (message, forum, thread, author, parent, created) VALUES($1, $2, $3, $4, $5, $6) " +
		"RETURNING id, author, created, forum, message, parent, thread, is_edited"
	GetPostsFlatPart1 = "SELECT p.id, p.author, p.created, p.forum, p.message, p.parent, p.thread, p.is_edited FROM posts p JOIN " +
		"(SELECT id FROM posts WHERE thread = $1 %s ORDER BY id %s LIMIT $2) AS sq ON sq.id = p.id "
	GetPostsFlatSincePart = "AND id %s $3 "
	GetPostsFlatPart2     = "ORDER BY id %s "
	GetPostsTree          = "SELECT p.id, p.author, p.created, p.forum, p.message, p.parent, p.thread, p.is_edited FROM posts p " +
		"JOIN (SELECT id FROM posts WHERE thread = $1 %s ORDER BY path %s LIMIT $2) AS sq ON sq.id = p.id ORDER BY path %s "
	GetPostsTreeSincePart = "AND path %s (SELECT path FROM posts WHERE id = $3) "
	GetPostsParentTree    = "SELECT p.id, p.author, p.created, p.forum, p.message, p.parent, p.thread, p.is_edited FROM posts p JOIN ( " +
		"SELECT id FROM posts WHERE parent = 0 AND thread = $1 %s ORDER BY id %s LIMIT $2) AS sq ON sq.id=p.path[1] "
	ParentTreeSincePart        = "AND id %s (SELECT path[1] FROM posts WHERE id=$3)"
	GetPostsParentTreePart2Alt = "ORDER BY path[1] %s, path "
)

func (db *DB) GetPost(postId string) (models.Post, int) {
	post := models.Post{}
	row := db.db.QueryRow(GetPost, postId)
	timeStamp := time.Time{}
	err := row.Scan(&post.ID, &post.Author, &timeStamp,
		&post.Forum, &post.Message, &post.Parent,
		&post.Thread, &post.IsEdited)
	if err == pgx.ErrNoRows {
		return post, EmptyResult
	}
	post.Created = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
	if err != nil {
		return post, DBError
	}
	return post, OK
}

func (db *DB) GetPostInfo(postId string, related []string) (models.PostFull, int) {
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
	authors := make([]string, 0)
	_, err = tx.Prepare("insert_posts", InsertPost)
	for _, post := range posts {
		ifUserExist, err := IsUserExist(tx, post.Author)
		if post.Parent != 0 {
			ifParentExist := false
			err := tx.QueryRow("SELECT true FROM POSTS WHERE id = $1 AND thread = $2", post.Parent, threadId).Scan(&ifParentExist)
			if err == pgx.ErrNoRows {
				return nil, Conflict
			}
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
		timeStamp := time.Time{}
		err = tx.QueryRow("insert_posts", post.Message, forumId, threadId, post.Author, post.Parent,
			timeString).Scan(&curPost.ID, &curPost.Author, &timeStamp,
			&curPost.Forum, &curPost.Message, &curPost.Parent, &curPost.Thread, &curPost.IsEdited)
		if err != nil {
			return nil, DBError
		}
		curPost.Created = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
		postsToReturn = append(postsToReturn, &curPost)
		authors = append(authors, curPost.Author)
	}
	postsLen := len(postsToReturn)
	if postsLen > 0 {
		_, err = tx.Exec("UPDATE forum SET posts_count = posts_count + $1 WHERE slug = $2", len(postsToReturn), postsToReturn[0].Forum)
		if err != nil {
			return nil, DBError
		}
		_, err := tx.Prepare("insert_authors", "INSERT INTO forum_to_users(forum, user_nick) VALUES ($1, $2) ON CONFLICT DO NOTHING;")
		if err != nil {
			return nil, DBError
		}
		for _, author := range authors {
			_, err = tx.Exec("insert_authors", forumId, author)
			if err != nil {
				return nil, DBError
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, DBError
	}

	return postsToReturn, OK
}

func (db *DB) CreatePostsById(id string, posts models.Posts) (models.Posts, int) {
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
	authors := make([]string, 0)
	_, err = tx.Prepare("insert_posts", InsertPost)
	for _, post := range posts {
		ifUserExist, err := IsUserExist(tx, post.Author)
		if err != nil {
			return nil, DBError
		}
		if post.Parent != 0 {
			ifParentExist := false
			err := tx.QueryRow("SELECT true FROM POSTS WHERE id = $1 AND thread = $2", post.Parent, id).Scan(&ifParentExist)
			if err == pgx.ErrNoRows {
				return nil, Conflict
			}
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
		timeStamp := time.Time{}
		err = tx.QueryRow("insert_posts", post.Message, forumId, id, post.Author, post.Parent,
			timeString).Scan(&curPost.ID, &curPost.Author, &timeStamp,
			&curPost.Forum, &curPost.Message, &curPost.Parent, &curPost.Thread, &curPost.IsEdited)
		if err != nil {
			return nil, DBError
		}
		curPost.Created = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
		postsToReturn = append(postsToReturn, &curPost)
		authors = append(authors, curPost.Author)
	}
	postsLen := len(postsToReturn)
	if postsLen > 0 {
		_, err = tx.Exec("UPDATE forum SET posts_count = posts_count + $1 WHERE slug = $2", len(postsToReturn), postsToReturn[0].Forum)
		if err != nil {
			return nil, DBError
		}
		_, err := tx.Prepare("insert_authors", "INSERT INTO forum_to_users(forum, user_nick) VALUES ($1, $2) ON CONFLICT DO NOTHING;")
		if err != nil {
			return nil, DBError
		}
		for _, author := range authors {
			_, err = tx.Exec("insert_authors", forumId, author)
			if err != nil {
				return nil, DBError
			}
		}
	}
	err = tx.Commit()
	if err != nil {
		return nil, DBError
	}
	return postsToReturn, OK
}

func (db *DB) GetPostsBySlug(slug string, limit string, since string,
	sort string, desc string) (models.Posts, int) {
	id := 0
	row := db.db.QueryRow(getThreadIdBySlug, slug)
	err := row.Scan(&id)
	if err == pgx.ErrNoRows {
		return nil, EmptyResult
	}
	if err != nil {
		return nil, DBError
	}
	switch sort {
	case "flat":
		return db.GetPostsFlat(strconv.Itoa(id), limit, since, desc)
	case "tree":
		return db.GetPostsTree(strconv.Itoa(id), limit, since, desc)
	case "parent_tree":
		return db.GetPostsParentTree(strconv.Itoa(id), limit, since, desc)
	}
	return models.Posts{}, OK
}

func (db *DB) GetPostsById(id string, limit string, since string,
	sort string, desc string) (models.Posts, int) {
	ifThreadExists := false
	err := db.db.QueryRow("SELECT true FROM threads WHERE id = $1", id).Scan(&ifThreadExists)
	if err == pgx.ErrNoRows {
		return models.Posts{}, EmptyResult
	}
	if err != nil {
		return nil, DBError
	}
	if !ifThreadExists {
		return models.Posts{}, EmptyResult
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

func (db *DB) GetPostsFlat(id string, limit string, since string, desc string) (models.Posts, int) {
	ifDesc, _ := strconv.ParseBool(desc)
	strDesc := "ASC"
	if ifDesc {
		strDesc = "DESC"
	}
	rows := &pgx.Rows{}
	err := errors.New("")
	if since != "" {
		actualSince := ""
		if ifDesc {
			actualSince = fmt.Sprintf(GetPostsFlatSincePart, "<")
		} else {
			actualSince = fmt.Sprintf(GetPostsFlatSincePart, ">")
		}
		query := fmt.Sprintf(GetPostsFlatPart1, actualSince, strDesc) + GetPostsFlatPart2
		rows, err = db.db.Query(fmt.Sprintf(query, strDesc), id, limit, since)
	} else {
		query := fmt.Sprintf(GetPostsFlatPart1, "", strDesc) + GetPostsFlatPart2
		rows, err = db.db.Query(fmt.Sprintf(query, strDesc), id, limit)
	}
	if err != nil {
		return nil, DBError
	}
	defer rows.Close()
	posts := make(models.Posts, 0)
	for rows.Next() {
		post := new(models.Post)
		timeStamp := time.Time{}
		err := rows.Scan(&post.ID, &post.Author, &timeStamp, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		post.Created = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
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
	rows := &pgx.Rows{}
	err := errors.New("")
	query := ""
	if since != "" {
		actualSince := ""
		if ifDesc {
			actualSince = fmt.Sprintf(GetPostsTreeSincePart, "<")
		} else {
			actualSince = fmt.Sprintf(GetPostsTreeSincePart, ">")
		}
		query = fmt.Sprintf(GetPostsTree, actualSince, strDesc, strDesc)
		rows, err = db.db.Query(query, id, limit, since)
	} else {
		query = fmt.Sprintf(GetPostsTree, "", strDesc, strDesc)
		rows, err = db.db.Query(query, id, limit)
	}
	if err != nil {
		return nil, DBError
	}
	defer rows.Close()
	posts := make(models.Posts, 0)
	for rows.Next() {
		post := new(models.Post)
		timeStamp := time.Time{}
		err := rows.Scan(&post.ID, &post.Author, &timeStamp, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		post.Created = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
		posts = append(posts, post)
	}
	return posts, OK
}

func (db *DB) GetPostsParentTree(id string, limit string, since string, desc string) (models.Posts, int) {
	ifDesc, _ := strconv.ParseBool(desc) // mb check error
	rows := &pgx.Rows{}
	err := errors.New("")
	strDesc := "ASC"
	if ifDesc {
		strDesc = "DESC"
	}
	query := ""
	if since != "" {
		actualSince := ""
		if ifDesc {
			actualSince = fmt.Sprintf(ParentTreeSincePart, "<")
		} else {
			actualSince = fmt.Sprintf(ParentTreeSincePart, ">")
		}
		query = fmt.Sprintf(GetPostsParentTree, actualSince, strDesc) + fmt.Sprintf(GetPostsParentTreePart2Alt, strDesc)
		rows, err = db.db.Query(query, id, limit, since)
	} else {
		query = fmt.Sprintf(GetPostsParentTree, "", strDesc) + fmt.Sprintf(GetPostsParentTreePart2Alt, strDesc)
		rows, err = db.db.Query(query, id, limit)
	}
	if err != nil {
		return nil, DBError
	}
	defer rows.Close()
	posts := make(models.Posts, 0)
	for rows.Next() {
		post := new(models.Post)
		timeStamp := time.Time{}
		err := rows.Scan(&post.ID, &post.Author, &timeStamp, &post.Forum, &post.Message,
			&post.Parent, &post.Thread, &post.IsEdited)
		if err != nil {
			return models.Posts{}, DBError
		}
		post.Created = timeStamp.Format("2006-01-02T15:04:05.999999999Z07:00")
		posts = append(posts, post)
	}
	return posts, OK
}
