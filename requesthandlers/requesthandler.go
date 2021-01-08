package requesthandlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"pulseservice/config"
	"pulseservice/utils"
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
	// So we need tha APP ID, and the start and end date from the request
	appId, appidErr := utils.GetQueryParam(r, "appid")
	if appidErr != nil {
		utils.ReturnWithError(http.StatusBadRequest, "No appid provided", w)
		return
	}
	sdate, sdateErr := utils.GetQueryParam(r, "sdate")
	if sdateErr != nil {
		utils.ReturnWithError(http.StatusBadRequest, "No startdate provided", w)
		return
	}
	edate, edateErr := utils.GetQueryParam(r, "edate")
	if edateErr != nil {
		utils.ReturnWithError(http.StatusBadRequest, "No end date provided", w)
		return
	}
	// Now try to convert the provided start date and end date to golang date types
	StartTime, StartTimeParseError := time.Parse("2006-01-02T15:04:05Z", sdate)
	if StartTimeParseError != nil {
		StartTime, StartTimeParseError = time.Parse("2006-01-02 15:04:05", sdate)

	}
	if StartTimeParseError != nil {
		utils.ReturnWithError(http.StatusBadRequest, "Could not parse StartDate", w)
		return
	}
	EndTime, EndTimeParseError := time.Parse("2006-01-02T15:04:05Z", edate)
	if EndTimeParseError != nil {
		EndTime, EndTimeParseError = time.Parse("2006-01-02 15:04:05", edate)

	}
	if EndTimeParseError != nil {
		utils.ReturnWithError(http.StatusBadRequest, "Could not parse End Date", w)
		return
	}

	MessageRows, GetMessagesErr := config.GetMessageForAppBetweenTimes(appId, StartTime, EndTime) //many confusion
	if GetMessagesErr != nil {
		fmt.Println("Bad Request", GetMessagesErr)
		return
	}

	outdata, marshaleerr := json.MarshalIndent(MessageRows, " ", " ")
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
