package server

import (
	"encoding/json"
	"net/http"
)

func (env Env) ServeTest(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Content-Type", "application/json;charset=utf-8")
	res.Header().Set("Access-Control-Allow-Origin", "*")
	var quiz Quiz
	env.DB.First(&quiz, 1)
	quiz.QueryQuestions(env)
	quizJson, err := json.Marshal(quiz)
	if err != nil {
		http.Error(res, "Error encoding JSON", 500)
		return
	}
	res.Write(quizJson)
}
