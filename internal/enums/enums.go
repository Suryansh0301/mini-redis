package constants

type StatusCode string

const (
	StatusCodeSuccess      StatusCode = "success"
	StatusCodeError        StatusCode = "error"
	StatusCodeNeedMoreData StatusCode = "needMoreData"
)
