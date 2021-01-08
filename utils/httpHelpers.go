package utils

import (
	"encoding/json"
	"errors"
	"log"

	"net/http"
)

func GetQueryParam(r *http.Request, paramName string) (string, error) {
	keys, ok := r.URL.Query()[paramName]

	if !ok || len(keys[0]) < 1 {
		log.Println("Url Param 'key' is missing")
		return "", errors.New("parametet not provided")
	}

	// Query()["key"] will return an array of items,
	// we only want the single item.
	return keys[0], nil

}

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
