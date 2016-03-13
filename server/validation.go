package server

import "strings"

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

// CheckAnswer accepts radio and checkbox answers and verifies the answer
func (userAnswer UserChoiceAnswer) CheckAnswer(question Question, answer interface{}) UserAnswerer {
	if value, found := answer.(uint); found {
		userAnswer.ChoiceID = value
		userAnswer.Correct.Bool = question.isCorrectChoiceAnswer(userAnswer.ChoiceID)
		userAnswer.Correct.Valid = true
		return userAnswer
	}
	return nil
}

// Validate returns true if model's parameters have valid values
func (userAnswer UserChoiceAnswer) Validate() bool {
	return userAnswer.UserID > 0 && userAnswer.QuestionID > 0 && userAnswer.ChoiceID > 0
}

// IsCorrect checks whether the answer is correct
func (userAnswer UserChoiceAnswer) IsCorrect() bool {
	return userAnswer.Correct.Valid && userAnswer.Correct.Bool
}

// CheckAnswer accepts fillblank and textarea answers, converts them to appropriate types and verifies their values
// textarea value cannot be directly verified, it is done later
func (userAnswer UserTextAnswer) CheckAnswer(question Question, answer interface{}) UserAnswerer {
	if value, found := answer.(string); found {
		userAnswer.Value = strings.TrimSpace(value)
		if question.Type == "textarea" {
			userAnswer.Correct.Bool = false
			userAnswer.Correct.Valid = false
		} else {
			userAnswer.Correct.Bool = question.isCorrectStringAnswer(userAnswer.Value)
			userAnswer.Correct.Valid = true
		}
		return userAnswer
	}
	return nil
}

// Validate returns true if model's parameters have valid values
func (userAnswer UserTextAnswer) Validate() bool {
	return userAnswer.UserID > 0 && userAnswer.QuestionID > 0 && len(userAnswer.Value) > 0
}

// IsCorrect checks whether the answer is correct
func (userAnswer UserTextAnswer) IsCorrect() bool {
	return userAnswer.Correct.Valid && userAnswer.Correct.Bool
}
