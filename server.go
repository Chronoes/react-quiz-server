package main

import (
	"github.com/Chronoes/ReactQuizServer/server"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

func main() {
	env := new(server.Env)
	err := env.InitDB("postgres", "user=vesikonna dbname=quiztest sslmode=disable")
	if err != nil {
		log.Fatal(err)
		return
	}
	// http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/test", env.ServeTest)
	log.Println("Listening on 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
