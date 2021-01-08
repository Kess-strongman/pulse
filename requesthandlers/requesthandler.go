package requesthandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pulseservice/config"
	"time"
)

func lala() {
	var malarkey time.Time
	fmt.Println("Woop", malarkey)

}

// ServiceGetStatusHandler - receives a request for service status
func ServiceGetStatusHandler(w http.ResponseWriter, r *http.Request) {
	// First we check that the request is valid - Don't worry at this stage we will implement this next week

	// Then once it's authorised - We get the service status from the Database

}

func ServiceGetLatestHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := config.GetLatestEntries()
	if err != nil {
		// Figure out how to return an error - status bad request
		fmt.Println("Bad Request", err)
		return
	}
	outdata, marshaleerr := json.MarshalIndent(rows, " ", " ")
	if marshaleerr == nil {
		w.Write(outdata)
	}
}

func ServiceGetMessageForAppBetweenTimesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err = config.GetMessageForAppBetweenTimes("pulseTest", 2021-01-07 13:53:24.86751, 2021-01-07 13:56:25.022257) //many confusion
	if err != nil {
		fmt.Println("Bad Request", err)
		return
	}
	outdata, marshaleerr := json.MarshalIndent(rows, " ", " ")
	if marshaleerr == nil {
		w.Write(outdata)
	}
}

func ServiceGetLatestMessageForAppHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := config.GetLatestMessageForApp("pulseTest")
	if err != nil {
		fmt.Println("Bad Request", err)
		return
	}
	outdata, marshaleerr := json.MarshalIndent(rows, " ", " ")
	if marshaleerr == nil {
		w.Write(outdata)
	}
}

func GetLatestServiceStatusMessagesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := config.GetLatestServiceStatusMessages()
	if err != nil {
		fmt.Println("Bad Request", err)
		return
	}
	outdata, marshaleerr := json.MarshalIndent(rows, " ", " ")
	if marshaleerr == nil {
		w.Write(outdata)
	}
}

func GetLatestHelloMessagesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := config.GetLatestHelloMessages()
	if err != nil {
		fmt.Println("Bad Request", err)
		return
	}
	outdata, marshaleerr := json.MarshalIndent(rows, " ", " ")
	if marshaleerr == nil {
		w.Write(outdata)
	}
}

// ServiceGetAlertHandler - receives a request for service alerts
func ServiceGetAlertHandler(w http.ResponseWriter, r *http.Request) {
	/*
		var ServiceName string
		var TS time.Time
		var Level string
		var Text string
		var Data string
		var EntryType string

		sentServiceName := r.URL.Query()["servicename"]
		sentTS := r.URL.Query()["ts"]*/
}
