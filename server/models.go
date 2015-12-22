package server

import "time"

type Choice struct {
	QuestionID uint
	Value      string
}

type Question struct {
	ID            uint     `json:"-" gorm:"primary_key"`
	QuizID        uint     `json:"-"`
	Type          string   `json:"type"`
	Question      string   `json:"question"`
	Choices       []Choice `json:"-"`
	ChoiceStrings []string `json:"choices" sql:"-"`
}

func (q *Question) StringifyChoices() {
	for _, choice := range q.Choices {
		q.ChoiceStrings = append(q.ChoiceStrings, choice.Value)
	}
}

type Quiz struct {
	ID        uint       `json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	Status    string     `json:"-"`
	Title     string     `json:"title"`
	Questions []Question `json:"questions"`
}

func (quiz *Quiz) QueryQuestions(env Env) {
	env.DB.Model(&quiz).Related(&quiz.Questions)
	for i, question := range quiz.Questions {
		switch question.Type {
		case "radio":
			fallthrough
		case "checkbox":
			env.DB.Model(&question).Related(&question.Choices)
			question.StringifyChoices()
			quiz.Questions[i] = question
		}
	}
}
