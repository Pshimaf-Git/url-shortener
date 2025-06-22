package resp

// Response is a structure that includes common server response fields.
type Response struct {
	Status string `json:"status"`          // should be `OK` or `error` only
	Error  string `json:"error,omitempty"` // error message
}

const (
	StatusOK    = "OK"
	StatusError = "error"
)

// OK return a responce wirh empty error and status equal `OK`
func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

// Error return a responce with given error and status equal `error`
func Error(err error) Response {
	return Response{
		Status: StatusError,
		Error:  err.Error(),
	}
}
