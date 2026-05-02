package constants

type ResultType string

const (
	ResultTypeSuccess      ResultType = "success"
	ResultTypeError        ResultType = "error"
	ResultTypeNeedMoreData ResultType = "needMoreData"
)
