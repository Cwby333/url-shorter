package mainresponse

type Response struct {
	Status string `json:"status" validate:"required"`
	Error  string `json:"error,omitempty"`
}

const (
	statusOK    = "OK"
	statusError = "Error"
)

func NewOK() Response {
	return Response{
		Status: statusOK,
	}
}

func NewError(errors ...string) Response {
	str := ""

	for _, err := range errors {
		str += err + " "
	}

	return Response{
		Status: statusError,
		Error:  str,
	}
}
