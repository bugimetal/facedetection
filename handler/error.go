package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/bugimetal/facedetection"

	"github.com/sirupsen/logrus"
)

// ErrorStatusCodes maps commonly returned errors to HTTP status codes
var ErrorStatusCodes = map[error]int{
	facedetection.ErrNoImageSpecified:      http.StatusBadRequest,
	facedetection.ErrBadInput:              http.StatusBadRequest,
	facedetection.ErrCantReadImage:         http.StatusBadRequest,
	facedetection.ErrImageTypeNotSupported: http.StatusBadRequest,
	facedetection.ErrNoFacesFound:          http.StatusOK,
}

// errorResponse represents error response structure
type errorResponse struct {
	Resource *errorResource `json:"error"`
}

type errorResource struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func newErrorResponse(err error) *errorResponse {
	return &errorResponse{Resource: &errorResource{Code: statusCode(err), Message: err.Error()}}
}

func (er *errorResource) Error() string {
	s := strings.SplitAfter(er.Message, ": ")
	msg := s[len(s)-1]

	return msg
}

// statusCode returns the HTTP status code that is appropriate for the specified error.
func statusCode(err error) int {
	if statusCode, ok := ErrorStatusCodes[err]; ok {
		return statusCode
	}

	return http.StatusInternalServerError
}

func (handler *Handler) Error(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(statusCode(err))

	if err := json.NewEncoder(w).Encode(newErrorResponse(err)); err != nil {
		logrus.Errorf("unable to decode struct to json: %s", err)
	}
}
