package server

import (
	"database/sql"
	"github.com/jinzhu/gorm"
	"strings"
	"time"
)

// QuestionChoice is the database model for questions that have multiple choices
type QuestionChoice struct {
	ID         uint   `json:"id" gorm:"primary_key"`
	QuestionID uint   `json:"-" gorm:"primary_key"`
	Value      string `json:"value" sql:"type:text"`
	IsAnswer   bool   `json:"-"`
}

// Question is the database model for quiz questions
type Question struct {
	ID       uint             `json:"id" gorm:"primary_key"`
	QuizID   uint             `json:"-"`
	Type     string           `json:"type"`
	Question string           `json:"question"`
	Choices  []QuestionChoice `json:"choices"`
}

func (question *Question) queryChoices(db gorm.DB) {
	db.Model(question).Related(&question.Choices)
}

func (question Question) isCorrectChoiceAnswer(answerID uint) bool {
	for _, choice := range question.Choices {
		if choice.IsAnswer && choice.ID == answerID {
			return true
		}
	}
	return false
}

func (question Question) isCorrectStringAnswer(userAnswer string) bool {
	userAnswer = strings.ToLower(userAnswer)
	for _, choice := range question.Choices {
		if userAnswer == strings.ToLower(choice.Value) {
			return true
		}
	}
	return false
}

// Quiz is the database model for defined quizzes
type Quiz struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	Status    string     `json:"-"`
	TimeLimit int        `json:"timeLimit"`
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

func (quiz *Quiz) queryQuestions(db gorm.DB) {
	db.Model(&quiz).Related(&quiz.Questions)
	for i, Question := range quiz.Questions {
		switch Question.Type {
		case "radio":
			fallthrough
		case "checkbox":
			db.Model(&Question).Related(&Question.Choices)
			quiz.Questions[i] = Question
		}
	}
}

// Useranswer defines an uniform way of saving different types of user answers
type Useranswer interface {
	save(Question, interface{}, chan<- Useranswer, chan<- bool)
	validate() bool
}

// UserChoiceAnswer is the database model for choice-based user answers
type UserChoiceAnswer struct {
	UserID     uint         `json:"-"`
	QuestionID uint         `json:"questionId"`
	ChoiceID   uint         `json:"choiceId"`
	Correct    sql.NullBool `json:"correct"`
}

func (userAnswer *UserChoiceAnswer) verifyAnswer(question Question, answer interface{}) bool {
	switch value := answer.(type) {
	case float64:
		userAnswer.ChoiceID = uint(value)
		userAnswer.Correct.Bool = question.isCorrectChoiceAnswer(userAnswer.ChoiceID)
		userAnswer.Correct.Valid = true
		return true
	default:
		userAnswer.QuestionID = 0
	}
	return false
}

func (userAnswer UserChoiceAnswer) save(question Question, answers interface{}, out chan<- Useranswer, correct chan<- bool) {
	isCorrect := true
	if question.Type == "radio" {
		if userAnswer.verifyAnswer(question, answers) {
			isCorrect = userAnswer.Correct.Bool
		}
		out <- userAnswer
	} else {
		for _, answer := range answers.([]interface{}) {
			if userAnswer.verifyAnswer(question, answer) {
				isCorrect = userAnswer.Correct.Bool
			}
			out <- userAnswer
		}
	}
	correct <- isCorrect
}

func (userAnswer UserChoiceAnswer) validate() bool {
	return userAnswer.UserID > 0 && userAnswer.QuestionID > 0 && userAnswer.ChoiceID > 0
}

// UserTextAnswer is the database model for text-based user answers
type UserTextAnswer struct {
	UserID     uint         `json:"-"`
	QuestionID uint         `json:"questionId"`
	Value      string       `json:"value" sql:"type:text"`
	Correct    sql.NullBool `json:"correct"`
}

func (userAnswer *UserTextAnswer) verifyAnswer(question Question, answer interface{}) bool {
	switch value := answer.(type) {
	case string:
		userAnswer.Value = strings.TrimSpace(value)
		if question.Type == "textarea" {
			userAnswer.Correct.Valid = false
		} else {
			userAnswer.Correct.Bool = question.isCorrectStringAnswer(userAnswer.Value)
			userAnswer.Correct.Valid = true
		}
		return true
	default:
		userAnswer.QuestionID = 0
	}
	return false
}

func (userAnswer UserTextAnswer) save(question Question, answers interface{}, out chan<- Useranswer, correct chan<- bool) {
	isCorrect := true
	if question.Type == "textarea" {
		if !userAnswer.verifyAnswer(question, answers) {
			isCorrect = false
		}
	} else {
		for _, answer := range answers.([]interface{}) {
			if userAnswer.verifyAnswer(question, answer) {
				isCorrect = userAnswer.Correct.Bool
			}
			out <- userAnswer
		}
	}
	correct <- isCorrect
}

func (userAnswer UserTextAnswer) validate() bool {
	return userAnswer.UserID > 0 && userAnswer.QuestionID > 0 && len(userAnswer.Value) > 0
}

// User is the database model for an anonymous user tied to a quiz
type User struct {
	ID        uint         `json:"id" gorm:"primary_key"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	Name      string       `json:"name"`
	Quiz      Quiz         `json:"-"`
	QuizID    uint         `json:"quizId"`
	TimeSpent uint         `json:"timeSpent"`
	Answers   []Useranswer `json:"answers"`
}
