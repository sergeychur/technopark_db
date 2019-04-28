package server

import (
	"github.com/go-chi/chi"
	"github.com/sergeychur/technopark_db/internal/models"
	"net/http"
)

func (serv *Server) CreateNewThreadPosts(w http.ResponseWriter, r *http.Request) {
	threadId := chi.URLParam(r, "slug_or_id")
	slugOrId := SlugOrId(threadId)
	posts := models.Posts{}
	err := ReadFromBody(r, w, &posts)
	if err != nil {
		return
	}
	stat := 0
	if slugOrId == slug {
		posts, stat = serv.db.CreatePostsBySlug(threadId, posts)
		DealCreateStatus(w, &posts, stat)
		return
	}
	if slugOrId == id {
		posts, stat = serv.db.CreatePostsById(threadId, posts)
		DealCreateStatus(w, &posts, stat)
		return
	}
	errText := models.Error{Message: "Invalid url"}
	WriteToResponse(w, http.StatusBadRequest, errText)
}

func (serv *Server) GetThreadInfo(w http.ResponseWriter, r *http.Request) {
	threadId := chi.URLParam(r, "slug_or_id")
	slugOrId := SlugOrId(threadId)
	thread := models.Thread{}
	stat := 0
	if slugOrId == slug {
		thread, stat = serv.db.GetThreadBySlug(threadId)
		DealGetStatus(w, &thread, stat)
		return
	}
	if slugOrId == id {
		thread, stat = serv.db.GetThreadById(threadId)
		DealGetStatus(w, &thread, stat)
		return
	}
	errText := models.Error{Message: "Invalid url"}
	WriteToResponse(w, http.StatusBadRequest, errText)
}

func (serv *Server) UpdateThread(w http.ResponseWriter, r *http.Request) {
	threadId := chi.URLParam(r, "slug_or_id")
	slugOrId := SlugOrId(threadId)

	threadUpdate := models.ThreadUpdate{}
	err := ReadFromBody(r, w, &threadUpdate)
	if err != nil {
		return
	}
	thread := models.Thread{}
	stat := 0
	if slugOrId == slug {
		thread, stat = serv.db.UpdateThreadBySlug(threadId, threadUpdate)
		DealGetStatus(w, &thread, stat)
		return
	}
	if slugOrId == id {
		thread, stat = serv.db.UpdateThreadById(threadId, threadUpdate)
		DealGetStatus(w, &thread, stat)
		return
	}
	errText := models.Error{Message: "Invalid url"}
	WriteToResponse(w, http.StatusBadRequest, errText)
}

func (serv *Server) GetThreadMessages(w http.ResponseWriter, r *http.Request) {
	threadId := chi.URLParam(r, "slug_or_id")
	slugOrId := SlugOrId(threadId)
	params := r.URL.Query()
	limits, ok := params["limit"]
	limit := ""
	if !ok {
		/*errText := models.Error{Message: "No limit"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return*/
		limit = "100"
	} else {
		limit = limits[0]
	}

	sinces, ok := params["since"]
	if !ok {
		errText := models.Error{Message: "No since"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return
	}
	since := sinces[0]

	sorts, ok := params["sort"]
	sort := ""
	if !ok {
		/*errText := models.Error{Message: "No sort"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return*/
		sort = "flat"
	} else {
		sort = sorts[0]
	}

	descs, ok := params["desc"]
	if !ok {
		errText := models.Error{Message: "No desc"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return
	}
	desc := descs[0]
	posts := models.Posts{}
	stat := 0
	if slugOrId == slug {
		posts, stat = serv.db.GetPostsBySlug(threadId, limit, since, sort, desc)
		DealGetStatus(w, &posts, stat)
		return
	}
	if slugOrId == id {
		posts, stat = serv.db.GetPostsById(threadId, limit, since, sort, desc)
		DealGetStatus(w, &posts, stat)
		return
	}
	errText := models.Error{Message: "Invalid url"}
	WriteToResponse(w, http.StatusBadRequest, errText)
}

func (serv *Server) Vote(w http.ResponseWriter, r *http.Request) {
	threadId := chi.URLParam(r, "slug_or_id")
	slugOrId := SlugOrId(threadId)

	vote := models.Vote{}
	err := ReadFromBody(r, w, &vote)
	if err != nil {
		return
	}

	if slugOrId == slug {
		thread, stat := serv.db.VoteBySlug(threadId, vote)
		DealGetStatus(w, &thread, stat)
		return
	}
	if slugOrId == id {
		thread, stat := serv.db.VoteById(threadId, vote)
		DealGetStatus(w, &thread, stat)
		return
	}

	errText := models.Error{Message: "Invalid url"}
	WriteToResponse(w, http.StatusBadRequest, errText)
}
