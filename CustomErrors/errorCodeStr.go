package CustomErrors

type GenericErrors interface {
	ErrorMsg() string
	Error() string
	ErrorCode() int
}

type GenericError struct {
	Code int
	Err  string
}

func NewGenericError(errorCode int, text string) GenericErrors {
	return &GenericError{errorCode, text}
}

func (e *GenericError) ErrorMsg() string {
	return e.Err
}

func (e *GenericError) Error() string {
	return e.ErrorMsg()
}

func (e *GenericError) ErrorCode() int {
	return e.Code
}
