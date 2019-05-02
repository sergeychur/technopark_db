package server

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/sergeychur/technopark_db/config"
	"github.com/sergeychur/technopark_db/internal/database"
	"log"
	"net/http"
	"os"
	"strconv"
)

type Server struct {
	router *chi.Mux
	db     *database.DB
	config *config.Config
}

func NewServer(pathToConfig string) (*Server, error) {
	server := new(Server)
	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	//r.Use(middleware.Recoverer)
	slugPattern := "^(\\d|\\w|-|_)*(\\w|-|_)(\\d|\\w|-|_)*$"
	idPattern := "^[0-9]+$"
	nickPattern := "^[A-Za-z0-9_\\.-]+$"

	subRouter := chi.NewRouter()
	subRouter.Post("/forum/create", server.CreateForum)
	subRouter.Post(fmt.Sprintf("/forum/{slug:%s}/create", slugPattern), server.CreateThread)
	subRouter.Get(fmt.Sprintf("/forum/{slug:%s}/details", slugPattern), server.GetForumInfo)
	subRouter.Get(fmt.Sprintf("/forum/{slug:%s}/threads", slugPattern), server.GetForumThreads)
	subRouter.Get(fmt.Sprintf("/forum/{slug:%s}/users", slugPattern), server.GetUsersByForum)

	subRouter.Get(fmt.Sprintf("/post/{id:%s}/details", idPattern), server.GetPostInfo)
	subRouter.Post(fmt.Sprintf("/post/{id:%s}/details", idPattern), server.EditPost)

	subRouter.Post("/service/clear", server.ClearDB)
	subRouter.Get("/service/status", server.GetDBInfo)

	subRouter.Post("/thread/{slug_or_id}/create", server.CreateNewThreadPosts)
	subRouter.Get("/thread/{slug_or_id}/details", server.GetThreadInfo)
	subRouter.Post("/thread/{slug_or_id}/details", server.UpdateThread)
	subRouter.Get("/thread/{slug_or_id}/posts", server.GetThreadMessages)
	subRouter.Post("/thread/{slug_or_id}/vote", server.Vote)

	subRouter.Post(fmt.Sprintf("/user/{nickname:%s}/create", nickPattern), server.CreateUser)
	subRouter.Get(fmt.Sprintf("/user/{nickname:%s}/profile", nickPattern), server.GetUserInfo)
	subRouter.Post(fmt.Sprintf("/user/{nickname:%s}/profile", nickPattern), server.UpdateUser)

	r.Mount("/api/", subRouter)
	server.router = r

	newConfig, err := config.NewConfig(pathToConfig)
	if err != nil {
		return nil, err
	}
	server.config = newConfig
	dbPort, err := strconv.Atoi(server.config.DBPort)
	if err != nil {
		return nil, err
	}
	db := database.NewDB(server.config.DBUser, server.config.DBPass,
		server.config.DBName, server.config.DBHost, uint16(dbPort))
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
