package CustomErrors

import "Concord/Structures"

func GenericErrorCodeHandler(gerr GenericErrors, response *Structures.Response) {
	response.Status = gerr.ErrorCode()
	if gerr.ErrorCode() > 4000 && gerr.ErrorCode() < 5000 {
		response.Msg = gerr.ErrorMsg()
	} else {
		response.Msg = "internal server error"
	}
}

func ErrorCodeHandler(errorCode int, err error, response *Structures.Response) {
	gerr := NewGenericError(errorCode, err.Error())
	GenericErrorCodeHandler(gerr, response)
}
