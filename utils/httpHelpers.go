package utils

import (
	"encoding/json"
	"log"

	"net/http"
)

// ReturnWithError notifies the user of an error with the supplied ErrorType and ErrorMessage
func ReturnWithError(ErrorType int, ErrorMessage string, w http.ResponseWriter) {
	type ret struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	log.Println("Return with Error called with message = ", ErrorMessage)
	var retval ret
	retval.Status = "error"
	retval.Message = ErrorMessage

	outbytes, outerr := json.Marshal(retval)
	if outerr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(outerr.Error()))

	} else {
		w.WriteHeader(ErrorType)
		w.Write(outbytes)
	}
}
