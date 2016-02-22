package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func (orm ORM) getActiveQuiz() (quiz *Quiz) {
	quiz = new(Quiz)
	orm.DB.Where("status = ?", "active").First(quiz)
	quiz.QueryQuestions(orm)
	return
}

func (orm ORM) addNewUser(name string, quizId uint) (user *User) {
	user = &User{Name: name, QuizID: quizId}
	orm.DB.Create(user)
	return
}

func (env Env) ServeTest(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	orm, err := env.OpenDB()
	if err != nil {
		http.Error(res, "Something happened", 500)
		log.Printf("Database opening failed: %v", err)
		return
	}
	defer orm.DB.Close()
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
	quiz := orm.getActiveQuiz()
	user := orm.addNewUser(userName, quiz.ID)
	quizJson, _ := json.Marshal(struct {
		*Quiz
		UserID uint `json:"userId"`
	}{
		Quiz:   quiz,
		UserID: user.ID,
	})
	res.Write(quizJson)
	log.Printf("Test ID %d served", quiz.ID)
}

type TestResults struct {
	UserID    uint
	TimeSpent uint
	Questions []struct {
		ID      uint
		Answers []interface{}
	}
}

func (results TestResults) parseAndSaveAnswers(orm ORM) (correctAnswers uint) {
	processedAnswers := make(chan UserAnswer)
	correct := make(chan bool)
	userAnswer := UserAnswer{UserID: results.UserID}
	maxAnswers := len(results.Questions)
	for _, question := range results.Questions {
		origQuestion := new(Question)
		orm.DB.First(origQuestion, question.ID)

		userAnswer.QuestionID = question.ID
		var action CheckAnswerFunc
		queryChoices := true
		switch origQuestion.Type {
		case "checkbox":
			fallthrough
		case "radio":
			action = userAnswer.checkByChoiceID
		case "fillblank":
			action = userAnswer.checkByString
		case "textarea":
			action = userAnswer.saveTextAreaAnswer
			queryChoices = false
		}

		if queryChoices {
			origQuestion.QueryChoices(orm)
		}

		maxAnswers += len(question.Answers)
		go action(*origQuestion, question.Answers, processedAnswers, correct)
	}

	correctAnswers = 0
	transact := orm.DB.Begin()
	for i := 0; i < maxAnswers; i++ {
		select {
		case userAnswer := <-processedAnswers:
			if userAnswer.QuestionID != 0 {
				transact.Create(&userAnswer)
			}
		case isCorrect := <-correct:
			if isCorrect {
				correctAnswers++
			}
		}
	}
	transact.Commit()
	close(processedAnswers)
	close(correct)
	return
}

func (results TestResults) saveUserTime(orm ORM) {
	orm.DB.First(&User{}, results.UserID).Update("time_spent", results.TimeSpent)
}

func (env Env) AcceptTest(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	orm, err := env.OpenDB()
	if err != nil {
		http.Error(res, "Something happened", 500)
		log.Printf("Database opening failed with %v", err)
		return
	}
	defer orm.DB.Close()
	dec := json.NewDecoder(req.Body)
	var results TestResults
	dec.Decode(&results)
	correctAnswers := results.parseAndSaveAnswers(*orm)
	results.saveUserTime(*orm)
	enc := json.NewEncoder(res)
	enc.Encode(map[string]interface{}{
		"correctAnswers": correctAnswers,
	})
}
