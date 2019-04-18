package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/sergeychur/technopark_db/config"
	"github.com/sergeychur/technopark_db/internal/database"
	"log"
	"net/http"
	"os"
)

type Server struct {
	router *chi.Mux
	db     *database.DB
	config *config.Config
}

func NewServer(pathToConfig string) (*Server, error) {
	server := new(Server)
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	slugPattern := "^(\\d|\\w|-|_)*(\\w|-|_)(\\d|\\w|-|_)*$"
	idPattern := "^[0-9]+$"
	nickPattern := "^[A-Za-z0-9_]+$"

	r.Post("/forum/create", server.CreateForum)
	r.Post(fmt.Sprintf("/forum/{slug:%s}/create", slugPattern), server.CreateThread)
	r.Get(fmt.Sprintf("/forum/{slug:%s}/details", slugPattern), server.GetForumInfo)
	r.Get(fmt.Sprintf("/forum/{slug:%s}/threads", slugPattern), server.GetForumThreads)
	r.Get(fmt.Sprintf("/forum/{slug:%s}/users", slugPattern), server.GetUsersByForum)

	r.Get(fmt.Sprintf("/post/{id:%s}/details", idPattern), server.GetPostInfo)
	r.Post(fmt.Sprintf("/post/{id:%s}/details", idPattern), server.EditPost)

	r.Post("/service/clear", server.ClearDB)
	r.Get("/service/status", server.GetDBInfo)

	r.Post("/thread/{slug_or_id}/create", server.CreateNewThreadPosts)
	r.Get("/thread/{slug_or_id}/details", server.GetThreadInfo)
	r.Post("/thread/{slug_or_id}/details", server.UpdateThread)
	r.Get("/thread/{slug_or_id}/posts", server.GetThreadMessages)
	r.Post("/thread/{slug_or_id}/vote", server.Vote)

	r.Post(fmt.Sprintf("/user/{nickname:%s}/create", nickPattern), server.CreateUser)
	r.Get(fmt.Sprintf("/user/{nickname:%s}/profile", nickPattern), server.GetUserInfo)
	r.Post(fmt.Sprintf("/user/{nickname:%s}/profile", nickPattern), server.UpdateUser)

	server.router = r

	newConfig, err := config.NewConfig(pathToConfig)
	if err != nil {
		return nil, err
	}
	server.config = newConfig
	db := database.NewDB(server.config.DBUser, server.config.DBPass,
		server.config.DBName, server.config.DBHost, server.config.DBPort)
	server.db = db
	return server, nil
}

func (serv *Server) Run() error {
	err := serv.db.Start()
	if err != nil {
		log.Printf("Failed to connect to DB: %s", err.Error())
		return err
	}
	defer serv.db.Close()
	port := serv.config.Port
	log.SetOutput(os.Stdout)
	log.Printf("Running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, serv.router))
	return nil
}
