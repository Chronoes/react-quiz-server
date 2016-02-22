package server

import (
	"database/sql"
	"strconv"
	"strings"
	"time"
)

type QuestionChoice struct {
	ID         uint   `json:"id" gorm:"primary_key"`
	QuestionID uint   `json:"-" gorm:"primary_key"`
	Value      string `json:"value" sql:"type:text"`
	IsAnswer   bool   `json:"-"`
}

type Question struct {
	ID       uint             `json:"id" gorm:"primary_key"`
	QuizID   uint             `json:"-"`
	Type     string           `json:"type"`
	Question string           `json:"question"`
	Choices  []QuestionChoice `json:"choices"`
}

func (question *Question) QueryChoices(orm ORM) {
	orm.DB.Model(question).Related(&question.Choices)
}

func (question Question) isCorrectMultiAnswer(answerId uint) bool {
	for _, choice := range question.Choices {
		if choice.IsAnswer && choice.ID == answerId {
			return true
		}
	}
	return false
}

func (question Question) isCorrectStringAnswer(userAnswer string) bool {
	for _, choice := range question.Choices {
		if userAnswer == choice.Value {
			return true
		}
	}
	return false
}

type Quiz struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	Status    string     `json:"-"`
	TimeLimit int        `json:"timeLimit"`
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

func (quiz *Quiz) QueryQuestions(orm ORM) {
	orm.DB.Model(&quiz).Related(&quiz.Questions)
	for i, question := range quiz.Questions {
		switch question.Type {
		case "radio":
			fallthrough
		case "checkbox":
			orm.DB.Model(&question).Related(&question.Choices)
			quiz.Questions[i] = question
		}
	}
}

type UserAnswer struct {
	UserID     uint         `json:"-"`
	QuestionID uint         `json:"questionId"`
	Value      string       `json:"value" sql:"type:text"`
	Correct    sql.NullBool `json:"correct"`
}

type CheckAnswerFunc func(Question, []interface{}, chan<- UserAnswer, chan<- bool)

func (userAnswer UserAnswer) checkByChoiceID(question Question, answers []interface{}, out chan<- UserAnswer, correct chan<- bool) {
	isCorrect := true
	for _, answer := range answers {
		switch value := answer.(type) {
		case float64:
			choiceId := uint(value)
			userAnswer.Correct.Bool = question.isCorrectMultiAnswer(choiceId)
			if !userAnswer.Correct.Bool {
				isCorrect = false
			}
			userAnswer.Correct.Valid = true
			userAnswer.Value = strconv.Itoa(int(choiceId))
			out <- userAnswer
		default:
			out <- UserAnswer{QuestionID: 0}
		}
	}
	correct <- isCorrect
}

func (userAnswer UserAnswer) checkByString(question Question, answers []interface{}, out chan<- UserAnswer, correct chan<- bool) {
	isCorrect := true
	for _, answer := range answers {
		blankAnswer := strings.TrimSpace(answer.(string))
		userAnswer.Correct.Bool = question.isCorrectStringAnswer(blankAnswer)
		if !userAnswer.Correct.Bool {
			isCorrect = false
		}
		userAnswer.Correct.Valid = true
		userAnswer.Value = blankAnswer
		out <- userAnswer
	}
	correct <- isCorrect
}

func (userAnswer UserAnswer) saveTextAreaAnswer(question Question, answers []interface{}, out chan<- UserAnswer, correct chan<- bool) {
	userAnswer.Value = strings.TrimSpace(answers[0].(string))
	userAnswer.Correct.Valid = false
	out <- userAnswer
	correct <- false
}

type User struct {
	ID        uint         `json:"id" gorm:"primary_key"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Name      string       `json:"name"`
	Quiz      Quiz         `json:"-"`
	QuizID    uint         `json:"quizId"`
	TimeSpent uint         `json:"timeSpent"`
	Answers   []UserAnswer `json:"answers"`
}
