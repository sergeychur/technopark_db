package server

import (
	//"encoding/json"
	"github.com/go-chi/chi"
	//"github.com/sergeychur/technopark_db/internal/database"
	"github.com/sergeychur/technopark_db/internal/models"
	//"io/ioutil"
	"net/http"
)

func (serv *Server) CreateForum(w http.ResponseWriter, r *http.Request) {
	forum := models.Forum{}
	err := ReadFromBody(r, w, &forum)
	if err != nil {
		return
	}
	forum, stat := serv.db.CreateForum(forum)
	DealCreateStatus(w, &forum, stat)
}

func (serv *Server) CreateThread(w http.ResponseWriter, r *http.Request) {
	forumId := chi.URLParam(r, "slug")
	thread := models.Thread{}
	err := ReadFromBody(r, w, &thread)
	if err != nil {
		return
	}
	thread, stat := serv.db.CreateThread(thread, forumId)
	DealCreateStatus(w, &thread, stat)
}

func (serv *Server) GetForumInfo(w http.ResponseWriter, r *http.Request) {
	forumId := chi.URLParam(r, "slug")
	forum := models.Forum{}
	forum, stat := serv.db.GetForum(forumId)
	DealGetStatus(w, &forum, stat)
}

func (serv *Server) GetForumThreads(w http.ResponseWriter, r *http.Request) {
	forumId := chi.URLParam(r, "slug")
	threads := models.Threads{}
	threads, stat := serv.db.GetForumThreads(forumId)
	DealGetStatus(w, &threads, stat)
}

func (serv *Server) GetUsersByForum(w http.ResponseWriter, r *http.Request) {
	forumId := chi.URLParam(r, "slug")
	users := models.Users{}
	users, stat := serv.db.GetForumUsers(forumId)
	DealGetStatus(w, &users, stat)
}
