package util

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	StatusCode int         `json:"status_code"`
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
}

func ResponseOK(w http.ResponseWriter, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), res.StatusCode)
	}
}

func ResponseError(w http.ResponseWriter, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), res.StatusCode)
	}
}
