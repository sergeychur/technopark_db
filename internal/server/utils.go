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
	noSince = fmt.Errorf("no since")
	noLimit = fmt.Errorf("no limit")
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

	if stat == database.Conflict {
		errText := models.Error{Message: "Conflict happened"}
		WriteToResponse(w, http.StatusConflict, errText)
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
		//errText := models.Error{Message: "No limit"}
		//WriteToResponse(w, http.StatusNotFound, errText)
		return noLimit
	}
	*limit = limits[0]
	if !idRegexp.MatchString(*limit) {
		errText := models.Error{Message: "Limit incorrect"}
		WriteToResponse(w, http.StatusNotFound, errText)
		return fmt.Errorf(errText.Message)
	}
	// TODO:(Me) parse date for correct

	descs, ok := params["desc"]
	if !ok {
		*desc = "asc"
	} else {
		*desc = descs[0]
		switch *desc {
		case "true":
			*desc = "desc"
		case "false":
			*desc = "asc"
		default:
			errText := models.Error{Message: "desc incorrect"}
			WriteToResponse(w, http.StatusBadRequest, errText)
			return fmt.Errorf(errText.Message)
		}
	}
	sinces, ok := params["since"]
	if !ok {
		return noSince
	}
	*since = sinces[0]
	return nil
}
