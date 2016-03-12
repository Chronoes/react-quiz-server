package server

import "testing"

func TestChoiceAnswer(t *testing.T) {
	question := Question{
		Choices: []QuestionChoice{
			QuestionChoice{ID: 1, IsAnswer: false},
			QuestionChoice{ID: 2, IsAnswer: true},
			QuestionChoice{ID: 3, IsAnswer: true},
			QuestionChoice{ID: 4, IsAnswer: false},
		},
	}

	if !question.isCorrectChoiceAnswer(2) {
		t.Error("Answer with ID 2 should be true but evaluated to false")
	}
	if question.isCorrectChoiceAnswer(4) {
		t.Error("Answer with ID 4 should be false but evaluated to true")
	}
}

func TestStringAnswer(t *testing.T) {
	question := Question{
		Choices: []QuestionChoice{
			QuestionChoice{Value: "Correct ANSWER"},
			QuestionChoice{Value: "EVEN MORE corRect ANswer"},
		},
	}

	if !question.isCorrectStringAnswer("even MORE correct Answer") {
		t.Error("Answer 'even MORE correct Answer' should be true but evaluated to false")
	}
	if question.isCorrectStringAnswer("wrong answer") {
		t.Error("Answer 'wrong answer' should be false but evaluated to true")
	}
}

func TestChoiceVerification(t *testing.T) {
	question := Question{
		Choices: []QuestionChoice{
			QuestionChoice{ID: 1, IsAnswer: false},
			QuestionChoice{ID: 2, IsAnswer: true},
			QuestionChoice{ID: 3, IsAnswer: true},
			QuestionChoice{ID: 4, IsAnswer: false},
		},
	}

	userAnswer := UserChoiceAnswer{}
	if !userAnswer.verifyAnswer(question, 3.0) {
		t.Error("Type casting has failed for valid answer type float")
	}

	if userAnswer.ChoiceID != 3 || !userAnswer.Correct.Valid {
		t.Error("Answer parameters were not set correctly")
	}

	userAnswerFail := UserChoiceAnswer{}
	if userAnswerFail.verifyAnswer(question, nil) {
		t.Error("Type casting failed for invalid answer type nil")
	}
}

func TestStringVerification(t *testing.T) {
	question := Question{
		Type: "fillblank",
		Choices: []QuestionChoice{
			QuestionChoice{Value: "Correct ANSWER"},
			QuestionChoice{Value: "EVEN MORE corRect ANswer"},
		},
	}

	userAnswer := UserTextAnswer{}
	answer := "fill me up"
	if !userAnswer.verifyAnswer(question, answer) {
		t.Error("Type casting has failed for valid answer type string")
	}

	if userAnswer.Value != answer || !userAnswer.Correct.Valid {
		t.Error("Answer parameters were not set correctly")
	}

	userAnswerFail := UserTextAnswer{}
	if userAnswerFail.verifyAnswer(question, 13) {
		t.Error("Type casting failed for invalid answer type int")
	}
}
