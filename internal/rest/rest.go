package rest

type ErrorResponse struct {
	Error string `json:"error"`
}

type RemovePersonRequest struct {
	Id int64 `json:"id"`
}
