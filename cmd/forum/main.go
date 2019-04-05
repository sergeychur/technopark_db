package main

import (
	"github.com/sergeychur/technopark_db/internal/server"
	"os"
)

func main() {
	pathToConfig := ""
	if len(os.Args) != 2 {
		panic("Usage: ./main <path_to_config>")
	} else {
		pathToConfig = os.Args[1]
	}
	serv, err := server.NewServer(pathToConfig)
	err = serv.Run()
	if err != nil {
		panic(err.Error())
	}
}
