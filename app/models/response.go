package models

type ErrorEnvelope struct {
	Error ErrorPayload `json:"error"`
}

type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Pagination struct {
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
}

type ListResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}
