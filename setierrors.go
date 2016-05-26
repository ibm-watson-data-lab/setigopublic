package main

import (
	"encoding/json"
	"net/http"
)

type SetiGoPublicError struct {
	Reason      string `'json:"reason"`
	ErrorString string `json:"error"`
}

func ReturnError(w http.ResponseWriter, error_code int, error_string string, reason string) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(error_code)
	var m = SetiGoPublicError{Reason: reason, ErrorString: error_string}
	if err := json.NewEncoder(w).Encode(m); err != nil {
		panic(err)
	}
}
