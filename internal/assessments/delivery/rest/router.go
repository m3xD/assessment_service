package rest

import "github.com/gorilla/mux"

type AssessmentRouter struct {
	assessmentHandler AssessmentHandler
}

func NewAssessmentRouter(assessmentHandler AssessmentHandler, router *mux.Router) *AssessmentRouter {
	return &AssessmentRouter{assessmentHandler: assessmentHandler}
}
