package main

import (
	"github.com/Chronoes/react-quiz-server/server"
	"github.com/julienschmidt/httprouter"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

func main() {
	env := server.Env{
		Production: os.Getenv("NODE_ENV") == "production",
	}
	err := env.InitDB("postgres", "user=vesikonna dbname=quiztest sslmode=disable")
	if err != nil {
		log.Fatal(err)
		return
	}
	router := httprouter.New()
	router.GET("/api/test", env.ApiDefaults(env.ServeTest))
	router.POST("/api/test", env.ApiDefaults(env.AcceptTest))
	router.NotFound = http.FileServer(http.Dir("static/"))
	log.Println("Listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
