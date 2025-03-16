package util

import (
	models "assessment_service/internal/model"
	"math/rand"

	"time"
)

// ShuffleQuestions randomizes the order of questions using Fisher-Yates shuffle algorithm
// This is used when an assessment has the RandomizeQuestions setting enabled
func ShuffleQuestions(questions []models.Question) []models.Question {
	// Create a copy of the original slice to avoid modifying the input
	result := make([]models.Question, len(questions))
	copy(result, questions)

	// Use current time as a seed for randomness
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Fisher-Yates shuffle algorithm
	for i := len(result) - 1; i > 0; i-- {
		// Generate a random index between 0 and i
		j := r.Intn(i + 1)

		// Swap elements at indices i and j
		result[i], result[j] = result[j], result[i]
	}

	return result
}

// ShuffleQuestionOptions randomizes the order of options for multiple-choice questions
// This can be used to further randomize the assessment experience
func ShuffleQuestionOptions(options []models.QuestionOption) []models.QuestionOption {
	// Create a copy of the original slice to avoid modifying the input
	result := make([]models.QuestionOption, len(options))
	copy(result, options)

	// Use current time as a seed for randomness
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Fisher-Yates shuffle algorithm
	for i := len(result) - 1; i > 0; i-- {
		// Generate a random index between 0 and i
		j := r.Intn(i + 1)

		// Swap elements at indices i and j
		result[i], result[j] = result[j], result[i]
	}

	return result
}
