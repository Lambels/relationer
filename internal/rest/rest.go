package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Lambels/relationer/internal"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type GetDepthResponse struct {
	Depth int `json:"depth"`
}

type idKey struct{}

func sendErrorResponse(w http.ResponseWriter, err error) {
	var ierr *internal.Error
	var errResp ErrorResponse
	var status int

	if !errors.As(err, &ierr) {
		errResp.Error = "internal server error"
		status = http.StatusInternalServerError
	} else {
		errResp.Error = ierr.Error()
		status = internal.StatusCodeFromECode(ierr.Code())
	}

	sendResponse(w, errResp, status)
}

func sendResponse(w http.ResponseWriter, res interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")

	payload, err := json.Marshal(res)
	if err != nil { // send json this way to separate encoding error (can call w.WriteHeader only once)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(status)
	w.Write(payload)
}
