package server

import (
	"github.com/go-chi/chi"
	"github.com/sergeychur/technopark_db/internal/models"
	"net/http"
)

func (serv *Server) GetPostInfo(w http.ResponseWriter, r *http.Request) {
	PostId := chi.URLParam(r, "slug")
	post := models.PostFull{}
	post, stat := serv.db.GetPostInfo(PostId)
	DealGetStatus(w, &post, stat)
}

func (serv *Server) EditPost(w http.ResponseWriter, r *http.Request) {
	PostId := chi.URLParam(r, "slug")
	postUpdate := models.PostUpdate{}
	err := ReadFromBody(r, w, &postUpdate)
	if err != nil {
		return
	}
	post, stat := serv.db.UpdatePost(PostId, postUpdate)
	DealGetStatus(w, &post, stat)
}
