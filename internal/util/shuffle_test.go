package util

import (
	models "assessment_service/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to check if two slices of Questions are identical in order and content
func areQuestionSlicesIdentical(a, b []models.Question) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		// Basic comparison based on ID, assuming IDs are unique and sufficient
		if a[i].ID != b[i].ID {
			return false
		}
	}
	return true
}

// Helper function to check if two slices of QuestionOptions are identical
func areOptionSlicesIdentical(a, b []models.QuestionOption) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		// Basic comparison based on ID
		if a[i].ID != b[i].ID {
			return false
		}
	}
	return true
}

// Helper function to check if slice b contains all elements of slice a (regardless of order)
func containsAllQuestions(a, b []models.Question) bool {
	if len(a) != len(b) {
		return false
	}
	mapA := make(map[uint]bool)
	for _, q := range a {
		mapA[q.ID] = true
	}
	for _, q := range b {
		if !mapA[q.ID] {
			return false // Element in b not found in a
		}
	}
	return true
}

func containsAllOptions(a, b []models.QuestionOption) bool {
	if len(a) != len(b) {
		return false
	}
	mapA := make(map[uint]bool)
	for _, o := range a {
		mapA[o.ID] = true
	}
	for _, o := range b {
		if !mapA[o.ID] {
			return false
		}
	}
	return true
}

func TestShuffleQuestions(t *testing.T) {
	originalQuestions := []models.Question{
		{ID: 1, Text: "Q1"},
		{ID: 2, Text: "Q2"},
		{ID: 3, Text: "Q3"},
		{ID: 4, Text: "Q4"},
		{ID: 5, Text: "Q5"},
	}

	// Test with multiple questions
	shuffled := ShuffleQuestions(originalQuestions)

	// 1. Check length
	assert.Equal(t, len(originalQuestions), len(shuffled), "Shuffled slice should have the same length")

	// 2. Check if all original elements are present
	assert.True(t, containsAllQuestions(originalQuestions, shuffled), "Shuffled slice must contain all original questions")

	// 3. Check if the original slice was modified (it shouldn't be)
	assert.True(t, areQuestionSlicesIdentical(originalQuestions, []models.Question{
		{ID: 1, Text: "Q1"}, {ID: 2, Text: "Q2"}, {ID: 3, Text: "Q3"}, {ID: 4, Text: "Q4"}, {ID: 5, Text: "Q5"},
	}), "Original slice should not be modified")

	// 4. Check for potential shuffling (probabilistic)
	// Run shuffle multiple times and check if the result is different from original at least once
	isDifferent := false
	for i := 0; i < 10; i++ { // Run a few times
		shuffledAgain := ShuffleQuestions(originalQuestions)
		if !areQuestionSlicesIdentical(originalQuestions, shuffledAgain) {
			isDifferent = true
			break
		}
	}
	// This assertion might occasionally fail due to pure chance, especially with small slices.
	// For robust testing of randomness, more sophisticated statistical tests would be needed.
	assert.True(t, isDifferent, "Shuffle function should likely produce a different order (probabilistic check)")

	// Test with empty slice
	emptyOriginal := []models.Question{}
	shuffledEmpty := ShuffleQuestions(emptyOriginal)
	assert.Empty(t, shuffledEmpty, "Shuffling an empty slice should result in an empty slice")

	// Test with single element slice
	singleOriginal := []models.Question{{ID: 10, Text: "Single"}}
	shuffledSingle := ShuffleQuestions(singleOriginal)
	assert.Equal(t, singleOriginal, shuffledSingle, "Shuffling a single-element slice should return the same slice")
	assert.True(t, areQuestionSlicesIdentical(singleOriginal, shuffledSingle))

}

func TestShuffleQuestionOptions(t *testing.T) {
	originalOptions := []models.QuestionOption{
		{ID: 1, OptionID: "a", Text: "Opt1"},
		{ID: 2, OptionID: "b", Text: "Opt2"},
		{ID: 3, OptionID: "c", Text: "Opt3"},
		{ID: 4, OptionID: "d", Text: "Opt4"},
	}

	shuffled := ShuffleQuestionOptions(originalOptions)

	// 1. Check length
	assert.Equal(t, len(originalOptions), len(shuffled), "Shuffled slice should have the same length")

	// 2. Check if all original elements are present
	assert.True(t, containsAllOptions(originalOptions, shuffled), "Shuffled slice must contain all original options")

	// 3. Check if original slice was modified
	assert.True(t, areOptionSlicesIdentical(originalOptions, []models.QuestionOption{
		{ID: 1, OptionID: "a", Text: "Opt1"}, {ID: 2, OptionID: "b", Text: "Opt2"}, {ID: 3, OptionID: "c", Text: "Opt3"}, {ID: 4, OptionID: "d", Text: "Opt4"},
	}), "Original slice should not be modified")

	// 4. Check for potential shuffling (probabilistic)
	isDifferent := false
	for i := 0; i < 10; i++ {
		shuffledAgain := ShuffleQuestionOptions(originalOptions)
		if !areOptionSlicesIdentical(originalOptions, shuffledAgain) {
			isDifferent = true
			break
		}
	}
	assert.True(t, isDifferent, "Shuffle function should likely produce a different order (probabilistic check)")

	// Test with empty slice
	emptyOriginal := []models.QuestionOption{}
	shuffledEmpty := ShuffleQuestionOptions(emptyOriginal)
	assert.Empty(t, shuffledEmpty, "Shuffling an empty slice should result in an empty slice")

	// Test with single element slice
	singleOriginal := []models.QuestionOption{{ID: 10, OptionID: "a", Text: "Single"}}
	shuffledSingle := ShuffleQuestionOptions(singleOriginal)
	assert.Equal(t, singleOriginal, shuffledSingle, "Shuffling a single-element slice should return the same slice")

}
