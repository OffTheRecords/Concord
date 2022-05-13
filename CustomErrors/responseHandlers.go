package CustomErrors

import "Concord/Structures"

func ErrorCodeHandler(gerr GenericErrors, response *Structures.Response) {
	response.Status = gerr.ErrorCode()
	if gerr.ErrorCode() > 4000 && gerr.ErrorCode() < 5000 {
		response.Msg = gerr.ErrorMsg()
	} else {
		response.Msg = "internal server error"
	}
}
