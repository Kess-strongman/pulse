package ingest

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"httpingest/config"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"httpingest/utils"
	//	"github.com/gorilla/handlers"
	//	"github.com/gorilla/mux"
)

/*

{
  "deviceID": "AIRSENSATEST001",
  "ts": "2020-04-24T14:00:01Z",
  "lat": 50,
  "lng": 1.2,
  "data": [
    {
      "valname": "PM10",
      "value": 22.34
    }
  ]
}

*/
// Payload carries the incoming Data
type Payload struct {
	DeviceID string                 `json:"deviceID"`
	TS       time.Time              `json:"ts"`
	Lat      float64                `json:"lat"`
	Lng      float64                `json:"lng"`
	Data     map[string]interface{} `json:"data"`
}

// Handler processes the incoming put request for data
func Handler(w http.ResponseWriter, r *http.Request) {
	// Check Method, headers, and content type
	contentType := r.Header.Get("Content-type")
	fmt.Println("ContentType = ", contentType)
	if contentType != "application/json" {
		utils.ReturnWithError(http.StatusBadRequest, "invalid content type: expected 'application/json'", w)
		return
	}
	// Look for the request token
	requestToken := r.Header.Get("X-Request-Token")
	if requestToken == "" {
		utils.ReturnWithError(http.StatusBadRequest, "missing request header expected 'X-Request-Token'", w)
		return
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error.
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination".
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	payloadString := string(b)
	fmt.Println(payloadString)
	var p Payload
	payloadErr := json.Unmarshal(b, &p)

	if payloadErr != nil {
		fmt.Println("ARGH", payloadErr.Error())
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxError):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset)
			http.Error(w, msg, http.StatusBadRequest)
		case errors.As(err, &unmarshalTypeError):
			msg := fmt.Sprintf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
			http.Error(w, msg, http.StatusBadRequest)
		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			msg := fmt.Sprintf("Request body contains unknown field %s", fieldName)
			http.Error(w, msg, http.StatusBadRequest)
		// Otherwise default to logging the error and sending a 500 Internal
		// Server Error response.
		default:
			log.Println(err.Error())
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
		return
	}
	SoughtDevice := config.Devices.GetDevice(p.DeviceID)
	if SoughtDevice.MAC == "" {
		log.Println(err.Error())
		http.Error(w, "could not find device", http.StatusNotFound)
		return
	}
	fmt.Println(SoughtDevice.Attributes)
	var deviceSecret string
	if val, ok := SoughtDevice.Attributes["devicesecret"]; ok {
		deviceSecret = val
	} else {
		http.Error(w, "could not determine device secret key", http.StatusBadRequest)
		return
	}
	data := []byte(string(b) + deviceSecret)

	hash := fmt.Sprintf("%x", md5.Sum(data))
	fmt.Println(hash)
	if hash != requestToken {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}
	fmt.Println(p.TS)
	if p.TS.IsZero() {
		p.TS = time.Now()

	}

	p.TS = p.TS.UTC()
	fmt.Println(p.TS)

	var DBData config.DBdatapayload
	DBData.Dstring = make(map[string]string)
	DBData.Dnum = make(map[string]float64)
	DBData.Dbool = make(map[string]bool)

	for k, dv := range p.Data {
		//	fmt.Println("Key =", k, " Value =", dv)
		switch dv.(type) {
		case int:
			// v is an int here, so e.g. v + 1 is possible.
			//fmt.Println("Integer: ", v)
			DBData.Dnum[k] = dv.(float64)
		case float64:
			// v is a float64 here, so e.g. v + 1.0 is possible.
			//		fmt.Println("Float64: ", v)
			DBData.Dnum[k] = dv.(float64)
		case string:
			// v is a string here, so e.g. v + " Yeah!" is possible.
			//		fmt.Println("String: ", v)
			DBData.Dstring[k] = dv.(string)
		case bool:
			// v is a string here, so e.g. v + " Yeah!" is possible.
			//		fmt.Println("Bool: ", v)
			DBData.Dbool[k] = dv.(bool)

		default:
			// And here I'm feeling dumb. ;)
			//		fmt.Printf("I don't know, ask stackoverflow.")
		}
	}
	if p.Lat == 0 && p.Lng == 0 {
		p.Lat = SoughtDevice.Lat
		p.Lng = SoughtDevice.Lng
	}
	fmt.Printf("Payload: %+v", p)
	fmt.Println(DBData)
	var newDBRow config.DBrowVersioned
	newDBRow.TS = p.TS
	newDBRow.LocID = SoughtDevice.LocID
	newDBRow.LocMAC = SoughtDevice.MAC
	newDBRow.Data = DBData
	newDBRow.Lat = p.Lat
	newDBRow.Lng = p.Lng

	var versionedRow config.DBRowVersion

	versionedRow.CreateTS = time.Now()
	versionedRow.Description = "Original Row"
	versionedRow.TransactionType = "Initial Insert"
	versionedRow.AuditReference = "raw data"
	versionedRow.NewRow = config.CopyVersionedDBRowToDBrow(newDBRow)
	newDBRow.VersionedData = append(newDBRow.VersionedData, versionedRow)
	fmt.Println(newDBRow.TS)

	result := PerformCalibrationCalculation(&newDBRow, SoughtDevice.Calibration)
	fmt.Println("Result : ", result)
	config.Workchan <- newDBRow
	fmt.Fprintf(w, "DBRow: %+v", newDBRow)

	/*	fmt.Println("Sending to Calib")

		DBRowToSave, caliberror := config.PerformCalibrationCalculation(SoughtDevice.Calibration, newDBRow)
		fmt.Println("BACK FROM CALIB", caliberror.Error())
		if caliberror == nil {
			fmt.Println(caliberror.Error)
			fmt.Println(DBRowToSave)
			config.Workchan <- DBRowToSave

		} else {
			config.Workchan <- newDBRow

		}
		fmt.Println("Pushed onto Channel")

		fmt.Fprintf(w, "DBRow: %+v", newDBRow)
	*/
}
func PerformCalibrationCalculation(dbrow *config.DBrowVersioned, calibrationrules config.CalibrationRules) error {
	fmt.Println("Calibration rules")
	fmt.Println(calibrationrules)
	if len(calibrationrules) != 0 && len(dbrow.VersionedData) > 0 {
		if dbrow.VersionedData[0].TransactionType == "Initial Insert" {
			fmt.Println("We have an initial Insert")
			var sourceData config.DBdatapayload
			sourceData.Dnum = make(map[string]float64)
			sourceData.Dbool = make(map[string]bool)
			sourceData.Dstring = make(map[string]string)
			for k, v := range dbrow.VersionedData[0].NewRow.Data.Dbool {
				sourceData.Dbool[k] = v
			}
			for k, v := range dbrow.VersionedData[0].NewRow.Data.Dnum {
				sourceData.Dnum[k] = v
			}
			for k, v := range dbrow.VersionedData[0].NewRow.Data.Dstring {
				sourceData.Dstring[k] = v
			}

			var calibratedrow config.DBRowVersion

			calibratedrow.CreateTS = time.Now()
			calibratedrow.Description = "CalibrationAppliedToOriginalInsert"
			calibratedrow.TransactionType = "Initial calibration adjustment"
			calibratedrow.AuditReference = "ingestion"
			transformationSpec, transformationerror := json.Marshal(calibrationrules)
			if transformationerror != nil {
				log.Println("Could not marshal transformation spec - storing blank string")
			}
			calibratedrow.Transformation = string(transformationSpec)

			for nk, nv := range sourceData.Dnum {
				fmt.Println(nk, " 0 ", nv)
				calibmodel := calibrationrules.GetCalibForPhenom(nk)
				if calibmodel.PhenomenonName != "" {
					if calibmodel.ModelName == "linearslopeandoffset" {
						//	fmt.Println("Doing calib for ", nk)
						//	fmt.Println("Original Valye = ", nv)
						if val, ok := calibmodel.CalibrationFactors["slope"]; ok {
							//fmt.Println("Slope = ", val)

							if val != 0 {

								nv = nv * val

							}
						}
						if val, ok := calibmodel.CalibrationFactors["offset"]; ok {
							//	fmt.Println("Offset = ", val)
							nv = nv + val

						}
						//fmt.Println("Calibrated Value = ", nv)
						sourceData.Dnum[nk] = nv

					}
				}
			}
			for k, _ := range dbrow.Data.Dnum {
				dbrow.Data.Dnum[k] = sourceData.Dnum[k]
			}

			//calibratedrow.NewRow = config.CopyVersionedDBRowToDBrow(dbrow)
			calibratedrow.NewRow.TS = dbrow.TS

			dbrow.VersionedData = append(dbrow.VersionedData, calibratedrow)

		}
	}

	fmt.Println("Trying to return from Calib")
	//	dbrow.Lat = 99
	return nil
}
