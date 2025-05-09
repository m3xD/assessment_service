/*
 * Secure Assessment Platform API
 *
 * API for managing secure online assessments with proctoring features
 *
 * API version: 1.0.0
 * Contact: support@secureassessment.example
 * Generated by: Swagger Codegen (https://github.com/swagger-api/swagger-codegen.git)
 */
package model

type Assessment struct {
	Questions []Question `json:"questions,omitempty"`

	Settings *AssessmentSettings `json:"settings,omitempty"`

	Id int64 `json:"id,omitempty"`

	Title string `json:"title,omitempty"`

	Subject string `json:"subject,omitempty"`

	Description string `json:"description,omitempty"`

	Duration int32 `json:"duration,omitempty"`

	Status string `json:"status,omitempty"`

	DueDate string `json:"dueDate,omitempty"`

	CreatedBy string `json:"createdBy,omitempty"`

	CreatedDate string `json:"createdDate,omitempty"`

	Attempts int32 `json:"attempts,omitempty"`

	PassingScore int32 `json:"passingScore,omitempty"`

	QuestionsCount int32 `json:"questionsCount,omitempty"`
}
