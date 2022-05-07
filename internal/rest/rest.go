package rest

type ErrorResponse struct {
	Error string `json:"error"`
}

type GetDepthResponse struct {
	Depth int `json:"depth"`
}

type RemovePersonRequest struct {
	Id int64 `json:"id"`
}

type GetDepthRequest struct {
	Id1 int64 `json:"id1"`
	Id2 int64 `json:"id2"`
}

type GetFriendshipRequest struct {
	Id int64 `json:"id"`
}
