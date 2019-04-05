package server

import (
	"github.com/sergeychur/technopark_db/internal/models"
	"net/http"
)

func (serv *Server) ClearDB(w http.ResponseWriter, r *http.Request) {
	err := serv.db.ClearDB()
	if err != nil {
		errorText := models.Error{Message: "error in database"}
		WriteToResponse(w, http.StatusInternalServerError, errorText)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (serv *Server) GetDBInfo(w http.ResponseWriter, r *http.Request) {
	status := models.Status{}
	status, stat := serv.db.GetDBInfo()
	DealGetStatus(w, &status, stat)
}
