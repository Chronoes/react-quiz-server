package server

import (
	"encoding/json"
	"github.com/julienschmidt/httprouter"
	"log"
	"net/http"
)

func (orm ORM) getActiveQuiz() (quiz *Quiz) {
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
	if err = req.ParseForm(); err != nil {
		http.Error(res, "Error parsing request", 400)
		log.Printf("Parsing request failed: %v", err)
		return
	}
	quiz := orm.getActiveQuiz()
	user := orm.addNewUser(req.FormValue("name"), quiz.ID)
	orm.DB.Close()
	quizJson, _ := json.Marshal(struct {
		*Quiz
		userId uint
	}{
		Quiz:   quiz,
		userId: user.ID,
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

func (orm ORM) parseAndSaveAnswers(results TestResults) (correctAnswers uint) {
	correctAnswers = 0
	for _, question := range results.Questions {
		var origQuestion Question
		orm.DB.First(&origQuestion, question.ID)
		origQuestion.QueryAnswers(orm)

		userAnswer := UserAnswer{UserID: results.UserID, QuestionID: question.ID}
		var action SaveAnswerFunc
		switch origQuestion.Type {
		case "checkbox":
			action = userAnswer.saveCheckboxAnswer
		case "radio":
			action = userAnswer.saveRadioAnswer
		case "fillblank":
			action = userAnswer.saveFillBlankAnswer
		case "textarea":
			action = userAnswer.saveTextAreaAnswer
		}

		if action(orm, origQuestion, question.Answers) {
			correctAnswers++
		}
	}
	return
}

func (env Env) AcceptTest(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	orm, err := env.OpenDB()
	if err != nil {
		http.Error(res, "Something happened", 500)
		log.Printf("Database opening failed with %v", err)
		return
	}
	dec := json.NewDecoder(req.Body)
	var results TestResults
	dec.Decode(&results)
	correctAnswers := orm.parseAndSaveAnswers(results)
	log.Println(results)
	orm.DB.Close()
	enc := json.NewEncoder(res)
	enc.Encode(map[string]interface{}{
		"correctAnswers": correctAnswers,
	})
}
