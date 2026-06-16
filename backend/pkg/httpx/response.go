// Package httpx provides small HTTP helpers: JSON responses and the structured
// error envelope mandated by Princípio XII ({code, message}).
package httpx

import (
	"encoding/json"
	"net/http"
)

// ErrorBody is the structured error envelope returned to clients.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON writes payload as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

// Error writes a structured error response. The message is user-facing (pt-BR)
// and must never contain secrets or stack traces (Princípio II/VI).
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorBody{Code: code, Message: message})
}
