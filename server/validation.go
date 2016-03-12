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

func (userAnswer *UserChoiceAnswer) verifyAnswer(question Question, answer interface{}) bool {
	value, found := answer.(float64)
	if found {
		userAnswer.ChoiceID = uint(value)
		userAnswer.Correct.Bool = question.isCorrectChoiceAnswer(userAnswer.ChoiceID)
		userAnswer.Correct.Valid = true
		return true
	}
	return false
}

// Save accepts radio and checkbox answers, converts them to appropriate types and verifies their values
func (userAnswer UserChoiceAnswer) Save(question Question, answers interface{}, out chan<- UserAnswerer, correct chan<- bool) {
	isCorrect := true
	if question.Type == "radio" {
		if userAnswer.verifyAnswer(question, answers) {
			isCorrect = userAnswer.Correct.Bool
		}
		out <- userAnswer
	} else {
		answersSlice, found := answers.([]interface{})
		if found {
			for _, answer := range answersSlice {
				if userAnswer.verifyAnswer(question, answer) {
					isCorrect = userAnswer.Correct.Bool
				}
				out <- userAnswer
			}
		} else {
			isCorrect = false
		}
	}
	correct <- isCorrect
}

// Validate returns true if model's parameters have valid values
func (userAnswer UserChoiceAnswer) Validate() bool {
	return userAnswer.UserID > 0 && userAnswer.QuestionID > 0 && userAnswer.ChoiceID > 0
}

func (userAnswer *UserTextAnswer) verifyAnswer(question Question, answer interface{}) bool {
	value, found := answer.(string)
	if found {
		userAnswer.Value = strings.TrimSpace(value)
		if question.Type == "textarea" {
			userAnswer.Correct.Valid = false
		} else {
			userAnswer.Correct.Bool = question.isCorrectStringAnswer(userAnswer.Value)
			userAnswer.Correct.Valid = true
		}
		return true
	}
	return false
}

// Save accepts fillblank and textarea answers, converts them to appropriate types and verifies their values
// textarea value cannot be directly verified, it is done later
func (userAnswer UserTextAnswer) Save(question Question, answers interface{}, out chan<- UserAnswerer, correct chan<- bool) {
	isCorrect := true
	if question.Type == "textarea" {
		if !userAnswer.verifyAnswer(question, answers) {
			isCorrect = false
		}
		out <- userAnswer
	} else {
		answersSlice, found := answers.([]interface{})
		if found {
			for _, answer := range answersSlice {
				if userAnswer.verifyAnswer(question, answer) {
					isCorrect = userAnswer.Correct.Bool
				}
				out <- userAnswer
			}
		} else {
			isCorrect = false
		}
	}
	correct <- isCorrect
}

// Validate returns true if model's parameters have valid values
func (userAnswer UserTextAnswer) Validate() bool {
	return userAnswer.UserID > 0 && userAnswer.QuestionID > 0 && len(userAnswer.Value) > 0
}
