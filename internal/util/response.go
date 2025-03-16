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

func ResponseMap(w http.ResponseWriter, data map[string]interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), statusCode)
	}
}

func ResponseInterface(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), statusCode)
	}
}

func ResponseError(w http.ResponseWriter, res Response) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(res.StatusCode)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), res.StatusCode)
	}
}
