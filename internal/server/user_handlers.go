package server

import (
	"github.com/go-chi/chi"
	"github.com/sergeychur/technopark_db/internal/models"
	"net/http"
)

func (serv *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	userNick := chi.URLParam(r, "nickname")
	user := models.User{}
	err := ReadFromBody(r, w, &user)
	if err != nil {
		return
	}
	user.Nickname = userNick
	user, stat := serv.db.CreateUser(user)
	DealCreateStatus(w, &user, stat)
}

func (serv *Server) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	userNick := chi.URLParam(r, "nickname")
	user := models.User{}
	user, stat := serv.db.GetUser(userNick)
	DealGetStatus(w, &user, stat)
}

func (serv *Server) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userNick := chi.URLParam(r, "nickname")
	userUpdate := models.UserUpdate{}
	err := ReadFromBody(r, w, &userUpdate)
	if err != nil {
		return
	}
	post, stat := serv.db.UpdateUser(userNick, userUpdate)
	DealGetStatus(w, &post, stat)
}
