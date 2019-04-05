package server

import (
	"encoding/json"
	"fmt"
	"github.com/sergeychur/technopark_db/internal/database"
	"github.com/sergeychur/technopark_db/internal/models"
	"io/ioutil"
	"net/http"
	"regexp"
)

const (
	noMatch     = -1
	slug    int = 1
	id      int = 2
)

var (
	slugRegExp *regexp.Regexp = regexp.MustCompile("^(\\d|\\w|-|_)*(\\w|-|_)(\\d|\\w|-|_)*$")
	idRegexp   *regexp.Regexp = regexp.MustCompile("^[0-9]+$")
)

func WriteToResponse(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)
	response, _ := json.Marshal(v)
	w.Write(response)
}

func ReadFromBody(r *http.Request, w http.ResponseWriter, v interface{}) error {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errText := models.Error{Message: "Cannot read body"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return fmt.Errorf(errText.Message)
	}
	err = json.Unmarshal(body, v)
	if err != nil {
		errText := models.Error{Message: "Cannot unmarshal json"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return fmt.Errorf(errText.Message)
	}
	return nil
}

func DealCreateStatus(w http.ResponseWriter, v interface{}, stat int) {
	if stat == database.DBError {
		errText := models.Error{Message: "Error in DB"}
		WriteToResponse(w, http.StatusInternalServerError, errText)
		return
	}
	if stat == database.OK {
		WriteToResponse(w, http.StatusCreated, v)
		return
	}

	if stat == database.EmptyResult {
		errText := models.Error{Message: "No item"}
		WriteToResponse(w, http.StatusNotFound, errText)
		return
	}

	if stat == database.Conflict {
		WriteToResponse(w, http.StatusConflict, v)
		return
	}
}

func DealGetStatus(w http.ResponseWriter, v interface{}, stat int) {
	if stat == database.DBError {
		errText := models.Error{Message: "Error in DB"}
		WriteToResponse(w, http.StatusInternalServerError, errText)
		return
	}
	if stat == database.OK {
		WriteToResponse(w, http.StatusOK, v)
		return
	}

	if stat == database.EmptyResult {
		errText := models.Error{Message: "No such item"}
		WriteToResponse(w, http.StatusNotFound, errText)
		return
	}
}

func SlugOrId(str string) int {
	if idRegexp.MatchString(str) {
		return id
	}
	if slugRegExp.MatchString(str) {
		return slug
	}
	return noMatch
}

func ParseParams(w http.ResponseWriter, r *http.Request,
	limit *string, since *string, desc *string) error {
	params := r.URL.Query()
	limits, ok := params["limit"]
	if !ok {
		errText := models.Error{Message: "No limit"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return fmt.Errorf(errText.Message)
	}
	*limit = limits[0]

	sinces, ok := params["since"]
	if !ok {
		errText := models.Error{Message: "No since"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return fmt.Errorf(errText.Message)
	}
	*since = sinces[0]

	descs, ok := params["desc"]
	if !ok {
		errText := models.Error{Message: "No desc"}
		WriteToResponse(w, http.StatusBadRequest, errText)
		return fmt.Errorf(errText.Message)
	}
	*desc = descs[0]
	return nil
}
