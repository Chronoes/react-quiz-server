package server

import (
	"encoding/json"
	"github.com/jinzhu/gorm"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func getActiveQuiz(db gorm.DB) (quiz *Quiz) {
	db.Where("status = ?", "active").First(quiz)
	quiz.queryQuestions(db)
	return
}

func addNewUser(db gorm.DB, name string, quizID uint) (user *User) {
	user.Name = name
	user.QuizID = quizID
	db.Create(user)
	return
}

// ServeQuiz serves currently active test to an user and registers their name to DB
func (env Env) ServeQuiz(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db, err := env.openDB()
	if err != nil {
		http.Error(res, "Something happened", 500)
		log.Printf("Database opening failed: %v", err)
		return
	}
	defer db.Close()
	if err = req.ParseForm(); err != nil {
		http.Error(res, "Error parsing request", 400)
		log.Printf("Parsing request failed: %v", err)
		return
	}
	userName := req.FormValue("name")
	if len(userName) == 0 {
		http.Error(res, "Required parameter name", 400)
		return
	}
	quiz := getActiveQuiz(db)
	user := addNewUser(db, userName, quiz.ID)
	quizJSON, _ := json.Marshal(struct {
		*Quiz
		UserID uint `json:"userId"`
	}{
		Quiz:   quiz,
		UserID: user.ID,
	})
	res.Write(quizJSON)
	log.Printf("Test ID %d served", quiz.ID)
}

type quizResults struct {
	UserID    uint
	TimeSpent uint
	Questions []struct {
		ID     uint
		Answer []interface{}
	}
}

func (results quizResults) parseAndSaveAnswers(db gorm.DB) (correctAnswers uint) {
	processedAnswers := make(chan Useranswer)
	correct := make(chan bool)
	maxAnswers := len(results.Questions)
	for _, question := range results.Questions {
		origQuestion := new(Question)
		db.First(origQuestion, question.ID)

		queryChoices := true
		var userAnswer Useranswer
		switch origQuestion.Type {
		case "checkbox":
			fallthrough
		case "radio":
			userAnswer = UserChoiceAnswer{UserID: results.UserID, QuestionID: question.ID}
		case "textarea":
			queryChoices = false
			fallthrough
		case "fillblank":
			userAnswer = UserTextAnswer{UserID: results.UserID, QuestionID: question.ID}
		}

		if queryChoices {
			origQuestion.queryChoices(db)
		}

		maxAnswers += len(question.Answer)
		go userAnswer.save(*origQuestion, question.Answer, processedAnswers, correct)
	}

	transact := db.Begin()
	for i := 0; i < maxAnswers; i++ {
		select {
		case userAnswer := <-processedAnswers:
			if userAnswer.validate() {
				transact.Create(&userAnswer)
			}
		case isCorrect := <-correct:
			if isCorrect {
				correctAnswers++
			}
		}
	}
	close(processedAnswers)
	close(correct)
	transact.Commit()
	return
}

func (results quizResults) saveUserTime(db gorm.DB) {
	db.First(&User{}, results.UserID).Update("time_spent", results.TimeSpent)
}

// AcceptQuizAnswers accepts a previously registered users answers to the quiz
func (env Env) AcceptQuizAnswers(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	db, err := env.openDB()
	if err != nil {
		http.Error(res, "Something happened", 500)
		log.Printf("Database opening failed with %v", err)
		return
	}
	defer db.Close()
	dec := json.NewDecoder(req.Body)
	var results quizResults
	dec.Decode(&results)
	correctAnswers := results.parseAndSaveAnswers(db)
	results.saveUserTime(db)
	enc := json.NewEncoder(res)
	enc.Encode(map[string]interface{}{
		"correctAnswers": correctAnswers,
	})
}
