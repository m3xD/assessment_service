package service

import (
	"assessment_service/internal/assessments/repository"
	repository2 "assessment_service/internal/attempts/repository"
	models "assessment_service/internal/model"
	repository3 "assessment_service/internal/questions/repository"
	repository4 "assessment_service/internal/users/repository"
	"assessment_service/internal/util"
	"errors"
	"fmt"

	"time"
)

type StudentService interface {
	GetAvailableAssessments(userID uint, params util.PaginationParams) ([]map[string]interface{}, int64, error)
	StartAssessment(userID, assessmentID uint) (*models.Attempt, []models.Question, *models.AssessmentSettings, *models.Assessment, error)
	GetAssessmentResultsHistory(userID, assessmentID uint) ([]map[string]interface{}, error)
	GetAttemptDetails(attemptID, userID uint) (*map[string]interface{}, error)
	SaveAnswer(attemptID, questionID uint, answer string, userID uint) error
	SubmitAssessment(attemptID, userID uint) (*map[string]interface{}, error)
	SubmitMonitorEvent(attemptID uint, eventType string, details map[string]interface{}, imageData []byte, userID uint) (*map[string]interface{}, error)
}

type studentService struct {
	assessmentRepo repository.AssessmentRepository
	attemptRepo    repository2.AttemptRepository
	questionRepo   repository3.QuestionRepository
	userRepo       repository4.UserRepository
}

func NewStudentService(
	assessmentRepo repository.AssessmentRepository,
	attemptRepo repository2.AttemptRepository,
	questionRepo repository3.QuestionRepository,
	userRepo repository4.UserRepository,
) StudentService {
	return &studentService{
		assessmentRepo: assessmentRepo,
		attemptRepo:    attemptRepo,
		questionRepo:   questionRepo,
		userRepo:       userRepo,
	}
}

func (s *studentService) GetAvailableAssessments(userID uint, params util.PaginationParams) ([]map[string]interface{}, int64, error) {
	// Check if user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, 0, errors.New("user not found")
	}

	return s.attemptRepo.FindAvailableAssessments(userID, params)
}

func (s *studentService) StartAssessment(userID, assessmentID uint) (*models.Attempt, []models.Question, *models.AssessmentSettings, *models.Assessment, error) {
	// Check if user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, nil, nil, nil, errors.New("user not found")
	}

	// Check if assessment exists and is active
	assessment, err := s.assessmentRepo.FindByID(assessmentID)
	if err != nil {
		return nil, nil, nil, nil, errors.New("assessment not found")
	}

	if assessment.Status != "published" {
		return nil, nil, nil, nil, errors.New("assessment is not active")
	}

	// Check if due date has passed
	if assessment.DueDate != nil && assessment.DueDate.Before(time.Now()) {
		return nil, nil, nil, nil, errors.New("assessment due date has passed")
	}

	// Check if user is taking assessment

	// Check if user has remaining attempts
	if assessment.Settings.AllowRetake {
		attemptsCount, err := s.attemptRepo.CountAttemptsByUserAndAssessment(userID, assessmentID)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if attemptsCount >= assessment.Settings.MaxAttempts {
			return nil, nil, nil, nil, errors.New("maximum attempts reached")
		}
	} else {
		// Check if user already completed this assessment
		hasCompleted, err := s.attemptRepo.HasCompletedAssessment(userID, assessmentID)
		if err != nil {
			return nil, nil, nil, nil, err
		}

		if hasCompleted {
			return nil, nil, nil, nil, errors.New("you have already completed this assessment")
		}
	}

	// Create new attempt
	attempt := &models.Attempt{
		UserID:       userID,
		AssessmentID: assessmentID,
		StartedAt:    time.Now(),
		Status:       "In Progress",
	}

	err = s.attemptRepo.Create(attempt)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// Get questions
	questions, err := s.questionRepo.FindByAssessmentID(assessmentID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	// If randomize is enabled, shuffle the questions
	if assessment.Settings.RandomizeQuestions {
		questions = util.ShuffleQuestions(questions)
	}

	// Remove correct answers for student view
	studentQuestions := make([]models.Question, len(questions))
	for i, q := range questions {
		studentQuestions[i] = models.Question{
			ID:      q.ID,
			Type:    q.Type,
			Text:    q.Text,
			Options: q.Options,
			Points:  q.Points,
		}
	}

	return attempt, studentQuestions, &assessment.Settings, assessment, nil
}

func (s *studentService) GetAssessmentResultsHistory(userID, assessmentID uint) ([]map[string]interface{}, error) {
	// Check if user exists
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Check if assessment exists
	_, err = s.assessmentRepo.FindByID(assessmentID)
	if err != nil {
		return nil, errors.New("assessment not found")
	}

	return s.attemptRepo.FindCompletedAttemptsByUserAndAssessment(userID, assessmentID)
}

func (s *studentService) GetAttemptDetails(attemptID, userID uint) (*map[string]interface{}, error) {
	// Get attempt details
	attempt, err := s.attemptRepo.FindByID(attemptID)
	if err != nil {
		return nil, errors.New("attempt not found")
	}

	// Check if attempt belongs to user
	if attempt.UserID != userID {
		return nil, errors.New("unauthorized access to attempt")
	}

	// Get assessment details
	assessment, err := s.assessmentRepo.FindByID(attempt.AssessmentID)
	if err != nil {
		return nil, errors.New("assessment not found")
	}

	// Calculate time remaining
	timeRemaining := int64(0)
	if attempt.Status == "In Progress" {
		endTime := attempt.StartedAt.Add(time.Duration(assessment.Duration) * time.Minute)
		if time.Now().Before(endTime) {
			timeRemaining = int64(endTime.Sub(time.Now()).Seconds())
		}
	}

	// Calculate progress
	totalQuestions := len(assessment.Questions)
	answeredQuestions := len(attempt.Answers)
	percentage := 0
	if totalQuestions > 0 {
		percentage = (answeredQuestions * 100) / totalQuestions
	}

	// Create response
	result := map[string]interface{}{
		"attemptId":     attempt.ID,
		"assessmentId":  attempt.AssessmentID,
		"title":         assessment.Title,
		"status":        attempt.Status,
		"startedAt":     attempt.StartedAt,
		"endsAt":        attempt.StartedAt.Add(time.Duration(assessment.Duration) * time.Minute),
		"timeRemaining": timeRemaining,
		"progress": map[string]interface{}{
			"answered":   answeredQuestions,
			"total":      totalQuestions,
			"percentage": percentage,
		},
		"answers": attempt.Answers,
	}

	return &result, nil
}

func (s *studentService) SaveAnswer(attemptID, questionID uint, answer string, userID uint) error {
	// Check if attempt exists
	attempt, err := s.attemptRepo.FindByID(attemptID)
	if err != nil {
		return errors.New("attempt not found")
	}

	// Check if attempt belongs to user
	if attempt.UserID != userID {
		return errors.New("unauthorized access to attempt")
	}

	// Check if attempt is still in progress
	if attempt.Status != "In Progress" {
		return errors.New("attempt is not in progress")
	}

	// Check if question exists and belongs to the assessment
	question, err := s.questionRepo.FindByID(questionID)
	if err != nil {
		return errors.New("question not found")
	}

	if question.AssessmentID != attempt.AssessmentID {
		return errors.New("question does not belong to this assessment")
	}

	// Check if answer is valid for the question type
	var isCorrect *bool
	switch question.Type {
	case "multiple-choice":
		// Check if answer is one of the options
		valid := false
		for _, option := range question.Options {
			if option.OptionID == answer {
				valid = true
				break
			}
		}

		if !valid {
			return errors.New("invalid answer for multiple-choice question")
		}

		correctVal := answer == question.CorrectAnswer
		isCorrect = &correctVal

	case "true-false":
		if answer != "true" && answer != "false" {
			return errors.New("answer for true-false question must be 'true' or 'false'")
		}

		correctVal := answer == question.CorrectAnswer
		isCorrect = &correctVal

	case "essay":
		// Essay answers are graded manually, so isCorrect is nil
		isCorrect = nil
	}

	// Check if the answer already exists
	existingAnswer, err := s.attemptRepo.FindAnswerByAttemptAndQuestion(attemptID, questionID)
	if err == nil && existingAnswer != nil {
		// Update existing answer
		existingAnswer.Answer = answer
		existingAnswer.IsCorrect = isCorrect
		return s.attemptRepo.UpdateAnswer(existingAnswer)
	}

	// Create new answer
	answerObj := &models.Answer{
		AttemptID:  attemptID,
		QuestionID: questionID,
		Answer:     answer,
		IsCorrect:  isCorrect,
	}

	return s.attemptRepo.SaveAnswer(answerObj)
}

func (s *studentService) SubmitAssessment(attemptID, userID uint) (*map[string]interface{}, error) {
	// Check if attempt exists
	attempt, err := s.attemptRepo.FindByID(attemptID)
	if err != nil {
		return nil, errors.New("attempt not found")
	}

	// Check if attempt belongs to user
	if attempt.UserID != userID {
		return nil, errors.New("unauthorized access to attempt")
	}

	// Check if attempt is still in progress
	if attempt.Status != "In Progress" {
		return nil, errors.New("attempt is not in progress")
	}

	// Get assessment details
	assessment, err := s.assessmentRepo.FindByID(attempt.AssessmentID)
	if err != nil {
		return nil, errors.New("assessment not found")
	}

	// Get all questions for this assessment
	questions, err := s.questionRepo.FindByAssessmentID(assessment.ID)
	if err != nil {
		return nil, err
	}

	// Calculate score
	totalQuestions := len(questions)
	correctAnswers := 0
	incorrectAnswers := 0
	unanswered := 0
	essayQuestions := 0
	totalPoints := 0.0
	earnedPoints := 0.0

	for _, q := range questions {
		totalPoints += q.Points

		answer, err := s.attemptRepo.FindAnswerByAttemptAndQuestion(attemptID, q.ID)
		if err != nil || answer == nil {
			unanswered++
			continue
		}

		if q.Type == "essay" {
			essayQuestions++
		} else if answer.IsCorrect != nil {
			if *answer.IsCorrect {
				correctAnswers++
				earnedPoints += q.Points
			} else {
				incorrectAnswers++
			}
		}
	}

	// Calculate percentage score
	var score float64
	if totalPoints > 0 {
		score = (earnedPoints / totalPoints) * 100
	}

	// Determine pass/fail status
	status := "Failed"
	if score >= assessment.PassingScore {
		status = "Passed"
	}

	// Calculate duration
	now := time.Now()
	duration := int(now.Sub(attempt.StartedAt).Minutes())

	// Generate feedback
	feedback := "Thank you for completing the assessment."
	if essayQuestions > 0 {
		feedback += " Your essay will be graded manually."
	}

	// Update attempt
	attempt.SubmittedAt = &now
	attempt.EndedAt = &now
	attempt.Score = &score
	attempt.Duration = &duration
	attempt.Status = status

	err = s.attemptRepo.Update(attempt)
	if err != nil {
		return nil, err
	}

	// Create response
	result := map[string]interface{}{
		"attemptId":    attempt.ID,
		"assessmentId": attempt.AssessmentID,
		"completed":    true,
		"submittedAt":  now,
		"duration":     duration,
		"showResults":  assessment.Settings.ShowResults,
		"results": map[string]interface{}{
			"score":            score,
			"totalQuestions":   totalQuestions,
			"correctAnswers":   correctAnswers,
			"incorrectAnswers": incorrectAnswers,
			"unanswered":       unanswered,
			"essayQuestions":   essayQuestions,
			"status":           status,
			"feedback":         feedback,
		},
	}

	return &result, nil
}

func (s *studentService) SubmitMonitorEvent(attemptID uint, eventType string, details map[string]interface{}, imageData []byte, userID uint) (*map[string]interface{}, error) {
	// Check if attempt exists and belongs to user
	attempt, err := s.attemptRepo.FindByID(attemptID)
	if err != nil {
		return nil, errors.New("attempt not found")
	}

	if attempt.UserID != userID {
		return nil, errors.New("unauthorized access to attempt")
	}

	// Check if attempt is still in progress
	if attempt.Status != "In Progress" {
		return nil, errors.New("attempt is not in progress")
	}

	// Determine severity based on event type
	severity := "NONE"
	message := "Event recorded."

	switch eventType {
	case "FACE_NOT_DETECTED":
		severity = "WARNING"
		message = "Please ensure your face is visible in the webcam at all times."
	case "MULTIPLE_FACES":
		severity = "CRITICAL"
		message = "Multiple faces detected. This is not allowed."
	case "LOOKING_AWAY":
		severity = "WARNING"
		message = "Please focus on your screen."
	case "SUSPICIOUS_OBJECT":
		severity = "WARNING"
		message = "Suspicious object detected. Please remove it."
	case "VOICE_DETECTED":
		severity = "WARNING"
		message = "Please remain quiet during the assessment."
	case "TAB_SWITCH":
		severity = "CRITICAL"
		message = "Switching tabs is not allowed during the assessment."
	}

	// Create suspicious activity record
	suspiciousActivity := &models.SuspiciousActivity{
		UserID:       userID,
		AssessmentID: attempt.AssessmentID,
		AttemptID:    attemptID,
		Type:         eventType,
		Details:      eventTypeToDetails(eventType, details),
		Timestamp:    time.Now(),
		Severity:     severity,
		ImageData:    imageData,
	}

	err = s.attemptRepo.SaveSuspiciousActivity(suspiciousActivity)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{
		"received": true,
		"severity": severity,
		"message":  message,
	}

	return &result, nil
}

// Helper function to convert event type and details to a string
func eventTypeToDetails(eventType string, details map[string]interface{}) string {
	switch eventType {
	case "FACE_NOT_DETECTED":
		return fmt.Sprintf("Face not detected for %.1f seconds (confidence: %.2f)",
			details["duration"].(float64), details["confidence"].(float64))
	case "MULTIPLE_FACES":
		return fmt.Sprintf("Multiple faces detected: %d",
			int(details["count"].(float64)))
	case "LOOKING_AWAY":
		return fmt.Sprintf("Looking away for %.1f seconds",
			details["duration"].(float64))
	case "TAB_SWITCH":
		return "User switched tabs"
	default:
		return fmt.Sprintf("%s detected", eventType)
	}
}
