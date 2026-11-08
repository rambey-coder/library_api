package utils

import (
	"encoding/json"
	"fmt"
	"library_api/internal/apperror"
	"log"
	"net/http"
	"strconv"
)

type Response[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func WriteJSON[T any](w http.ResponseWriter, status int, msg string, data T, err *Error) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if status == http.StatusNoContent {
		return nil
	}

	return json.NewEncoder(w).Encode(Response[T]{
		Success: status >= http.StatusOK && status < http.StatusBadRequest,
		Message: msg,
		Data:    data,
		Error:   err,
	})
}

func SuccessResponse[T any](w http.ResponseWriter, status int, msg string, data T) error {

	return WriteJSON(w, status, msg, data, nil)
}

func ErrorResponse(w http.ResponseWriter, r *http.Request, code int, msg string) error {
	var requestID string
	if r != nil {
		if val := r.Context().Value(RequestIDKey); val != nil {
			if s, ok := val.(string); ok {
				requestID = s
			} else {
				requestID = fmt.Sprint(val)
			}
		}
	}

	err := &Error{
		Code:      http.StatusText(code),
		Message:   msg,
		RequestID: requestID,
	}

	return WriteJSON[any](w, code, msg, nil, err)
}

func QueryInt(r *http.Request, key string, def int) (int, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return def, nil
	}

	v, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s parameter", key)
	}
	return v, nil
}

func PathID(r *http.Request, key string) (int64, error) {
	v, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	if err != nil || v <= 0 {
		return 0, fmt.Errorf("invalid %s", key)
	}
	return v, nil
}

func ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	status := apperror.HTTPStatus(err)
	var message string

	if status >= http.StatusInternalServerError {
		log.Printf("Internal server error %v", err)
		message = "internal server error"
	} else {
		log.Printf("An error occurred %v", err)
		message = err.Error()
	}

	ErrorResponse(w, r, status, message)

}

func QueryBool(r *http.Request, key string, def bool) (bool, error) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return def, nil
	}

	v, err := strconv.ParseBool(val)
	if err != nil {
		return def, fmt.Errorf("invalid %s", key)
	}
	return v, nil
}
