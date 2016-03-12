package server

import (
	"database/sql"
	"github.com/jinzhu/gorm"
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

func (question *Question) queryChoices(db *gorm.DB) {
	db.Model(question).Related(&question.Choices)
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

func (quiz *Quiz) queryQuestions(db *gorm.DB) {
	db.Model(quiz).Related(&quiz.Questions)
	for i, question := range quiz.Questions {
		switch question.Type {
		case "radio":
			fallthrough
		case "checkbox":
			question.queryChoices(db)
			quiz.Questions[i] = question
		}
	}
}

// UserAnswerer defines an uniform way of saving different types of user answers
type UserAnswerer interface {
	Save(Question, interface{}, chan<- UserAnswerer, chan<- bool)
	Validate() bool
}

// UserChoiceAnswer is the database model for choice-based user answers
type UserChoiceAnswer struct {
	UserID     uint         `json:"-"`
	QuestionID uint         `json:"questionId"`
	ChoiceID   uint         `json:"choiceId"`
	Correct    sql.NullBool `json:"correct"`
}

// UserTextAnswer is the database model for text-based user answers
type UserTextAnswer struct {
	UserID     uint         `json:"-"`
	QuestionID uint         `json:"questionId"`
	Value      string       `json:"value" sql:"type:text"`
	Correct    sql.NullBool `json:"correct"`
}

// User is the database model for an anonymous user tied to a quiz
type User struct {
	ID        uint           `json:"id" gorm:"primary_key"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Name      string         `json:"name"`
	Quiz      Quiz           `json:"-"`
	QuizID    uint           `json:"quizId"`
	TimeSpent uint           `json:"timeSpent"`
	Answers   []UserAnswerer `json:"answers"`
}
