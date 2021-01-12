package ingestionhandlers

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"pulseservice/config"
	"pulseservice/utils"
	"strings"
	"time"
)

// HelloMessage
type HelloMessage struct {
	ServiceName string    `json:"serviceName"`
	TS          time.Time `json:"ts"`
}
type ServiceStatusMessage struct {
	ServiceName string `json:"serviceName"`
	DataPoionts map[string]interface{}
	TS          time.Time `json:"ts"`
}
type ServiceAlert struct {
	ServiceName string    `json:"serviceName"`
	TS          time.Time `json:"ts"`
	Level       int       `json:"alertLevel"`
	Text        string    `json:"text"`
	Data        string    `json:"data"`
	EntryType   string    `json:"entryType"`
}

func ServiceStatusHandler(w http.ResponseWriter, r *http.Request) {

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") == true {
		bodyBytes, BodyReadErr := ioutil.ReadAll(r.Body)
		if BodyReadErr != nil {
			utils.ReturnWithError(http.StatusInternalServerError, "could not read body", w)
			return
		}

		appid := r.Header.Get("X-APP-ID")
		Token := config.AA.GetTokenUsingAppid(appid)
		if Token != " " {
			NewHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%xs", bodyBytes)+Token)))
			if NewHash != r.Header.Get("X-KEY-SECRET") {
				log.Println("Hash error")
				utils.ReturnWithError(401, "unauthorized hashy banana", w)
				return
			}

		} else {
			utils.ReturnWithError(401, "unauthorized token", w)
			return
		}

		var incomingStatusMessage ServiceStatusMessage
		bodyBytesUnmarshalError := json.Unmarshal(bodyBytes, &incomingStatusMessage)
		if bodyBytesUnmarshalError != nil {
			utils.ReturnWithError(http.StatusInternalServerError, "could not unmarshal body", w)
			return

		}
		inserterror := config.InsertTimeSeriesRow(incomingStatusMessage.ToDBrow())
		if inserterror != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(inserterror.Error()))

		} else {
			w.Write([]byte("ok"))

		}
	} else {
		utils.ReturnWithError(http.StatusBadRequest, "you must send JSON", w)
		return
	}
	w.Write([]byte("StatusSaved"))
}

func ServiceAlertHandler(w http.ResponseWriter, r *http.Request) {

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") == true {
		fmt.Println("We have some JSON")
		bodyBytes, BodyReadErr := ioutil.ReadAll(r.Body)
		if BodyReadErr != nil {
			utils.ReturnWithError(http.StatusInternalServerError, "could not read body", w)
			return
		}
		appid := r.Header.Get("X-APP-ID")
		Token := config.AA.GetTokenUsingAppid(appid)
		if Token != " " {
			NewHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%xs", bodyBytes)+Token)))
			if NewHash != r.Header.Get("X-KEY-SECRET") {
				log.Println("Hash error")
				utils.ReturnWithError(401, "unauthorized hash", w)
				return
			}

		} else {
			utils.ReturnWithError(401, "unauthorized token", w)
			return
		}

		var incomingAlertMessage ServiceAlert
		bodyBytesUnmarshalError := json.Unmarshal(bodyBytes, &incomingAlertMessage)
		if bodyBytesUnmarshalError != nil {
			utils.ReturnWithError(http.StatusInternalServerError, "could not unmarshal body", w)
			return
		}
		config.InsertTimeSeriesRow(incomingAlertMessage.ToDBrow())
		w.Write([]byte("ok"))

	} else {
		utils.ReturnWithError(http.StatusBadRequest, "you must send JSON", w)
		return
	}
	w.Write([]byte("StatusSaved"))
}

func HelloHandler(w http.ResponseWriter, r *http.Request) {

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") == true {
		fmt.Println("We have some JSON")
		bodyBytes, BodyReadErr := ioutil.ReadAll(r.Body)
		if BodyReadErr != nil {
			utils.ReturnWithError(http.StatusInternalServerError, "could not read body", w)
			return
		}
		appid := r.Header.Get("X-APP-ID")
		Token := config.AA.GetTokenUsingAppid(appid)
		if Token != " " {
			NewHash := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%xs", bodyBytes)+Token)))
			if NewHash != r.Header.Get("X-KEY-SECRET") {
				log.Println("Hash error: Received :", r.Header.Get("X-KEY-SECRET"), " Calculated ", NewHash)
				utils.ReturnWithError(401, "unauthorized hash", w)
				return
			}

		} else {

			utils.ReturnWithError(401, "unauthorized token", w)
			return
		}

		var incomingHelloMessage HelloMessage
		bodyBytesUnmarshalError := json.Unmarshal(bodyBytes, &incomingHelloMessage)
		if bodyBytesUnmarshalError != nil {
			utils.ReturnWithError(http.StatusInternalServerError, "could not unmarshal body", w)
			return
		}
		InsertError := config.InsertTimeSeriesRow(incomingHelloMessage.ToDBrow())
		if InsertError != nil {
			log.Println("Error saving hello message", InsertError.Error())
		}
		w.Write([]byte(incomingHelloMessage.ServiceName + "\n"))

	} else {
		utils.ReturnWithError(http.StatusBadRequest, "you must send JSON", w)
		return
	}
	w.Write([]byte("Hello!"))
}
func (me *HelloMessage) ToDBrow() config.DBrow {
	var tempro config.DBrow
	tempro.LocationID = me.ServiceName
	tempro.DeviceID = "HelloMessage"
	tempro.TS = me.TS.UTC()
	tempro.Lat = 50.9
	tempro.Lng = -1.5
	tempDstring := make(map[string]string)
	tempDstring["message"] = "Hello"
	tempro.Data.Dstring = tempDstring

	return tempro
}

func (me *ServiceAlert) ToDBrow() config.DBrow {
	var tempro config.DBrow
	tempro.LocationID = me.ServiceName
	tempro.DeviceID = "ServiceAlert"
	tempro.TS = me.TS.UTC()
	tempro.Lat = 50.9
	tempro.Lng = -1.5
	tempDstring := make(map[string]string)
	tempDstring["level"] = fmt.Sprintf("%d", me.Level)
	tempDstring["data"] = me.Data
	tempDstring["text"] = me.Text
	tempDstring["entryType"] = me.EntryType
	tempro.Data.Dstring = tempDstring

	return tempro
}

func (me *ServiceStatusMessage) ToDBrow() config.DBrow {
	var tempro config.DBrow
	tempro.LocationID = me.ServiceName
	tempro.DeviceID = "ServiceStatusMessage"
	tempro.TS = me.TS.UTC()
	tempro.Lat = 50.9
	tempro.Lng = -1.5
	tempDnum := make(map[string]float64)
	tempDstring := make(map[string]string)
	tempDbool := make(map[string]bool)
	for statusName, status := range me.DataPoionts {
		switch v := status.(type) {
		case float64:
			tempDnum[statusName] = v
		case int:
			tempDnum[statusName] = float64(v)
		case string:
			tempDstring[statusName] = v
		case bool:
			tempDbool[statusName] = v
		}
	}
	tempro.Data.Dnum = tempDnum
	tempro.Data.Dstring = tempDstring
	tempro.Data.Dbool = tempDbool

	return tempro
}
