package routes

import (
	"github.com/gorilla/mux"
	"net/http"
)

type OptionRouter struct {
	router *mux.Router
}

func NewOptionRouter(router *mux.Router) *OptionRouter {
	return &OptionRouter{router: router}
}

func (r *OptionRouter) OptionRegisterRouter() {
	r.router.PathPrefix("/api/v1/auth/login").Methods(http.MethodOptions).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.WriteHeader(http.StatusOK)
	})
}
