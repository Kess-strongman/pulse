package ingestionhandlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// HelloMessage
type HelloMessage struct {
	ServiceName string `json:"serviceName"`
}
type ServiceStatusMessage struct {
	ServiceName   string  `json:"serviceName"`
	MemoryUsageMB float64 `json:"memoryUsageMB"`
}
type ServiceAlert struct {
	ServiceName string    `json:"serviceName"`
	TimeStamp   time.Time `json:"ts"`
	Level       string    `json:"alertLevel"`
	Text        string    `json:"text"`
	Data        string    `json:"data"`
	EntryType   string    `json:"entryType"`
}

func ServiceStatusHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") == true {
		fmt.Println("We have some JSON")
		bodyBytes, BodyReadErr := ioutil.ReadAll(r.Body)
		if BodyReadErr != nil {
			log.Printf("Error reading body: %v", BodyReadErr)

			return
		}

		var incomingStatusMessage ServiceStatusMessage
		bodyBytesUnmarshalError := json.Unmarshal(bodyBytes, &incomingStatusMessage)
		if bodyBytesUnmarshalError != nil {
			fmt.Println("Error ", bodyBytesUnmarshalError.Error())
			w.Write([]byte("error"))
		}

		w.Write([]byte("ok"))
	} else {
		fmt.Println("We dont have JSON")
		w.Write([]byte("error"))
	}
}

func ServiceAlertHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") == true {
		fmt.Println("We have some JSON")
		bodyBytes, BodyReadErr := ioutil.ReadAll(r.Body)
		if BodyReadErr != nil {
			log.Printf("Error reading body: %v", BodyReadErr)

			return
		}

		var incomingAlertMessage ServiceAlert
		bodyBytesUnmarshalError := json.Unmarshal(bodyBytes, &incomingAlertMessage)
		if bodyBytesUnmarshalError != nil {
			fmt.Println("Error ", bodyBytesUnmarshalError.Error())
			w.Write([]byte("error"))
		}

		w.Write([]byte("ok"))

	} else {
		fmt.Println("We dont have JSON")
		w.Write([]byte("error"))
	}
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Method)

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") == true {
		fmt.Println("We have some JSON")
		bodyBytes, BodyReadErr := ioutil.ReadAll(r.Body)
		if BodyReadErr != nil {
			log.Printf("Error reading body: %v", BodyReadErr)

			return
		}

		var incomingHelloMessage HelloMessage
		bodyBytesUnmarshalError := json.Unmarshal(bodyBytes, &incomingHelloMessage)
		if bodyBytesUnmarshalError != nil {
			fmt.Println("Error ", bodyBytesUnmarshalError.Error())
		}

		w.Write([]byte(incomingHelloMessage.ServiceName + "\n"))

	} else {
		fmt.Println("We dont have JSON")
	}

	w.Write([]byte("Hello!"))
}
