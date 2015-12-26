package server

import (
	"strings"
	"time"
)

type QuestionChoice struct {
	ID         uint   `json:"id" gorm:"primary_key"`
	QuestionID uint   `json:"-" gorm:"primary_key"`
	Value      string `json:"value" sql:"type:text"`
}

type QuestionAnswer struct {
	QuestionID uint
	ChoiceID   uint
}

type Question struct {
	ID       uint             `json:"id" gorm:"primary_key"`
	QuizID   uint             `json:"-"`
	Type     string           `json:"type"`
	Question string           `json:"question"`
	Choices  []QuestionChoice `json:"choices"`
	Answers  []QuestionAnswer `json:"-"`
}

func (question *Question) QueryAnswers(orm ORM) {
	orm.DB.Model(question).Related(&question.Answers)
}

func (question Question) QueryAnswerChoices(orm ORM) (choices []QuestionChoice) {
	if len(question.Answers) > 0 {
		ids := make([]uint, len(question.Answers))
		for i, answer := range question.Answers {
			ids[i] = answer.ChoiceID
		}
		orm.DB.Where(ids).Find(&choices)
	}
	return
}

func (question Question) isCorrectMultiAnswer(answerId uint) bool {
	for _, answer := range question.Answers {
		if answer.ChoiceID == answerId {
			return true
		}
	}
	return false
}

func (question Question) isCorrectFillBlankAnswer(choices []QuestionChoice, userAnswer string) bool {
	for _, choice := range choices {
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
	UserID     uint   `json:"-"`
	QuestionID uint   `json:"questionId"`
	Value      string `json:"value" sql:"type:text"`
}

type SaveAnswerFunc func(ORM, Question, []interface{}) bool

func (userAnswer UserAnswer) saveCheckboxAnswer(orm ORM, question Question, answers []interface{}) (correct bool) {
	correct = true
	ids := make([]string, len(answers))
	for i, answer := range answers {
		choiceId := uint(answer.(float64))
		ids[i] = string(choiceId)
		if correct && !question.isCorrectMultiAnswer(choiceId) {
			correct = false
		}
	}
	userAnswer.Value = strings.Join(ids, ",")
	orm.DB.Create(&userAnswer)
	return
}

func (userAnswer UserAnswer) saveRadioAnswer(orm ORM, question Question, answers []interface{}) (correct bool) {
	value := uint(answers[0].(float64))
	correct = question.isCorrectMultiAnswer(value)
	userAnswer.Value = string(value)
	orm.DB.Create(&userAnswer)
	return
}

func (userAnswer UserAnswer) saveFillBlankAnswer(orm ORM, question Question, answers []interface{}) (correct bool) {
	correct = true
	choices := question.QueryAnswerChoices(orm)
	for _, answer := range answers {
		blankAnswer := strings.TrimSpace(answer.(string))
		if correct && !question.isCorrectFillBlankAnswer(choices, blankAnswer) {
			correct = false
		}
		userAnswer.Value = blankAnswer
		orm.DB.Create(&userAnswer)
	}
	return
}

func (userAnswer UserAnswer) saveTextAreaAnswer(orm ORM, question Question, answers []interface{}) bool {
	value := answers[0].(string)
	userAnswer.Value = value
	orm.DB.Create(&userAnswer)
	return false
}

type User struct {
	ID        uint         `json:"id" gorm:"primary_key"`
	CreatedAt time.Time    `json:"createdAt"`
	Name      string       `json:"name"`
	Quiz      Quiz         `json:"-"`
	QuizID    uint         `json:"quizId"`
	Answers   []UserAnswer `json:"answers"`
}
