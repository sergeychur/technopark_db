package server

import (
	"github.com/go-chi/chi"
	"github.com/sergeychur/technopark_db/internal/models"
	"net/http"
	"strings"
)

func (serv *Server) GetPostInfo(w http.ResponseWriter, r *http.Request) {
	PostId := chi.URLParam(r, "id")
	params := r.URL.Query()
	relatedStr, ok := params["related"]
	related := make([]string, 0)
	if ok {
		related = strings.Split(relatedStr[0], ",")
	}
	post := models.PostFull{}
	post, stat := serv.db.GetPostInfo(PostId, related)
	DealGetStatus(w, &post, stat)
}

func (serv *Server) EditPost(w http.ResponseWriter, r *http.Request) {
	PostId := chi.URLParam(r, "id")
	postUpdate := models.PostUpdate{}
	err := ReadFromBody(r, w, &postUpdate)
	if err != nil {
		return
	}
	post, stat := serv.db.UpdatePost(PostId, postUpdate)
	DealGetStatus(w, &post, stat)
}
