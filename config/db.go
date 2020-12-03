package config

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"sync"

	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // This is sql stuff - it's ok to hide it
)

// DB is the global DB handle
var DB *sql.DB

// DBerr is a gloal DB error
var DBerr error

// DBrow is a "standard" timeseries row
type DBrow struct {
	TS       time.Time
	LocID    string
	LocMAC   string
	Lat      float64
	Lng      float64
	Data     DBdatapayload
	CreateTS time.Time
}

// DBrowVersioned is a versioned timeseries row
type DBrowVersioned struct {
	TS            time.Time
	LocID         string
	LocMAC        string
	Lat           float64
	Lng           float64
	Data          DBdatapayload
	VersionedData DBRowVersionCollection
	CreateTS      time.Time
}

// DBRowVersionCollection is a collection of versioned DB Rows
type DBRowVersionCollection []DBRowVersion

// DBRowVersion is a the version element in a DBRowversion collection
type DBRowVersion struct {
	CreateTS         time.Time
	Description      string
	TransactionType  string
	CommittingUserID string
	NewRow           DBrow
	Transformation   string
	AuditReference   string
}

// DBdatapayload is the standard container for phenom data in a data row
type DBdatapayload struct {
	Dstring map[string]string
	Dnum    map[string]float64
	Dbool   map[string]bool
}

// CopyVersionedDBRowToDBrow copies a
func CopyVersionedDBRowToDBrow(inrow DBrowVersioned) DBrow {
	var outrow DBrow
	outrow.TS = inrow.TS
	outrow.LocID = inrow.LocID
	outrow.LocMAC = inrow.LocMAC
	outrow.Lat = inrow.Lat
	outrow.Lng = inrow.Lng
	outrow.Data.Dbool = make(map[string]bool)
	for k, v := range inrow.Data.Dbool {
		outrow.Data.Dbool[k] = v
	}
	outrow.Data.Dnum = make(map[string]float64)
	for k, v := range inrow.Data.Dnum {
		outrow.Data.Dnum[k] = v
	}
	outrow.Data.Dstring = make(map[string]string)
	for k, v := range inrow.Data.Dstring {
		outrow.Data.Dstring[k] = v
	}
	outrow.CreateTS = inrow.CreateTS
	return outrow
}

// CopyDBRow provides a proper copy of the DBrow and its maps
func CopyDBRow(inrow DBrow) DBrow {
	var outrow DBrow
	outrow.TS = inrow.TS
	outrow.LocID = inrow.LocID
	outrow.LocMAC = inrow.LocMAC
	outrow.Lat = inrow.Lat
	outrow.Lng = inrow.Lng
	outrow.Data.Dbool = make(map[string]bool)
	for k, v := range inrow.Data.Dbool {
		outrow.Data.Dbool[k] = v
	}
	outrow.Data.Dnum = make(map[string]float64)
	for k, v := range inrow.Data.Dnum {
		outrow.Data.Dnum[k] = v
	}
	outrow.Data.Dstring = make(map[string]string)
	for k, v := range inrow.Data.Dstring {
		outrow.Data.Dstring[k] = v
	}
	outrow.CreateTS = inrow.CreateTS
	return outrow
}

// InitDatabase initialises the database connection
func InitDatabase(connectionString string) {
	DB, DBerr = sql.Open("mysql", connectionString)
	fmt.Println("DB Connected")
	if DBerr != nil {
		fmt.Println(DBerr)
	}
	res := DB.Ping()
	if res != nil {
		fmt.Println(res)
	}
}

// InsertTimeSeriesRows inserts a set of versioned dbrows into the database
func InsertTimeSeriesRows(workitems []DBrowVersioned) error {
	log.Println("Insert Time Series Rows Called")
	if len(workitems) == 0 {
		log.Println("No items to work on")
		return nil
	}
	tablename := Config.GetTSTableName()
	valsholder := "(?,?,?,?,?,?,?), "
	var valscontainer []interface{}
	var sqlbuffer bytes.Buffer
	sqlbuffer.WriteString("REPLACE INTO " + tablename + " (ts, locid, locmac, lat, lng, data,versioned_data) VALUES ")
	for _, v := range workitems {

		valscontainer = append(valscontainer, v.TS.Format("2006-01-02T15:04:05.000Z07:00"))
		valscontainer = append(valscontainer, v.LocID)
		valscontainer = append(valscontainer, v.LocMAC)
		valscontainer = append(valscontainer, v.Lat)
		valscontainer = append(valscontainer, v.Lng)
		dataInBytes, err := json.Marshal(v.Data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)

		}
		valscontainer = append(valscontainer, string(dataInBytes))

		versionedDataInBytes, verr := json.MarshalIndent(v.VersionedData, " ", " ")
		if verr != nil {
			log.Println("error: ", verr)

		}
		valscontainer = append(valscontainer, string(versionedDataInBytes))

		sqlbuffer.WriteString(valsholder)

	}
	qry := sqlbuffer.String()

	// Snip the trailing comma from the query
	qry = qry[:len(qry)-2]

	_, err := DB.Exec(qry, valscontainer...)
	if err != nil {
		log.Println("Error saving workitems - need to figure out what to do here.", err)
		//fmt.Println(qry)
		//Deadlock found when trying to get lock; try restarting transaction
		return errors.New("Could not save workitems")
	}
	log.Println(len(workitems), " Workitems saved")

	return nil
}

// Workchan is the channel into which DBRows are placed by the injestion handler
var Workchan chan DBrowVersioned

// RunInjestionWorker reads new rows from the work channel and then inserts them into the database when either the number of rows to be saved hits 200 or when the
// PushToDBInterval passes
func RunInjestionWorker() {
	log.Println("Starting Ingestion worker")
	Workchan = make(chan DBrowVersioned)
	var workitems []DBrowVersioned

	defer WG.Done()
	run := true
	for run == true {
		timer := time.NewTimer(time.Second * time.Duration(Config.GetPushToDBInterval()))
		//	fmt.Println("Timer set for ", Config.GetPushToDBInterval(), " seconds")
		select {
		case <-Done:
			// If the done channel has something on it it's time to shut down
			log.Println("Close signal received for Injestion worker")
			if len(workitems) > 0 {
				InsertTimeSeriesRows(workitems)
				workitems = nil
			}
			run = false
			break
		case p := <-Workchan:
			//		fmt.Println("Fromw workchan")
			workitems = append(workitems, p)
			if len(workitems) >= Config.GetPushToDBMaxItemsInQueue() {
				fmt.Println("Saving ", len(workitems), " items")

				log.Println("Saving ", len(workitems), " items")
				InsertTimeSeriesRows(workitems)
				workitems = nil
			}
			break
		case <-timer.C:
			//		fmt.Println("Time Expired checking workitems - ", len(workitems))
			if len(workitems) > 0 {
				//			fmt.Println("Saving workitems")
				log.Println("Saving ", len(workitems), " items")
				InsertTimeSeriesRows(workitems)
				workitems = nil
			}
			break
		}
		if run != true {
			log.Println("Closing HTTP Injestion worker")
			break
		}

	}
}

type Device struct {
	MAC            string
	LocID          string
	Manufacturer   string
	Model          string
	Calibration    CalibrationRules
	Sensors        []Sensor
	Attributes     map[string]string
	Lat            float64
	Lng            float64
	LocationStatus string
}

func RunDeviceCacheWorker() {

	defer WG.Done()
	run := true
	Devices.LoadDevices()
	for run == true {
		fmt.Println("Loading devices")
		timer := time.NewTimer(time.Second * time.Duration(Config.GetDeviceCacheInterval()))
		select {
		case <-Done:
			// If the done channel has something on it it's time to shut down
			log.Println("Close signal received for Device Cache worker")
			run = false
			break
		case <-timer.C:
			Devices.LoadDevices()
			break
		}
		if run != true {
			log.Println("Closing Device Cache worker")
			break
		}

	}
}

type Sensor struct {
	Penomenon string `json:"phenomenon"`
	Units     string `json:"units"`
}
type CalibrationRules []CalibrationRule
type CalibrationRule struct {
	PhenomenonName     string `json:"phenomenon"`
	ModelName          string
	CalibrationFactors map[string]float64 `json:"calibrationFactor"`
}

type DeviceCollection struct {
	List map[string]Device
	sync.RWMutex
}

var Devices DeviceCollection

func (d *DeviceCollection) LoadDevices() {
	var tmpDeviceMap map[string]Device
	tmpDeviceMap = make(map[string]Device)
	LoadDeviceSQL := `SELECT  d.device_mac, d.device_locid, d.device_manuf,d.device_model, d.calibrationvalues, d.device_sensors,d.deviceattributes, l.location_lat, l.location_lng, l.status
	FROM aq_devices d 
	LEFT JOIN aq_locations l ON d.device_locid = l.loc_id
	WHERE d.transport = "HTTPINGEST"`
	rows, err := DB.Query(LoadDeviceSQL)
	if err != nil {
		log.Println("Error trying to device list from DB:", err.Error())
		return
	}
	defer rows.Close()
	for rows.Next() {
		var dbMac sql.NullString
		var dbLocid sql.NullString
		var dbManufacturer sql.NullString
		var dbModel sql.NullString
		var dbCalibrationValues sql.NullString
		var dbDeviceSensors sql.NullString
		var dbDeviceAttributes sql.NullString
		var dbLat sql.NullFloat64
		var dbLng sql.NullFloat64
		var dbStatus sql.NullString
		err := rows.Scan(&dbMac, &dbLocid, &dbManufacturer, &dbModel, &dbCalibrationValues, &dbDeviceSensors, &dbDeviceAttributes, &dbLat, &dbLng, &dbStatus)
		if err == nil {
			if dbMac.Valid && dbLocid.Valid && dbManufacturer.Valid && dbModel.Valid && dbDeviceSensors.Valid {
				var tmpDevice Device
				tmpDevice.LocID = dbLocid.String
				tmpDevice.MAC = dbMac.String
				tmpDevice.Manufacturer = dbManufacturer.String
				tmpDevice.Model = dbModel.String
				var sensorColl []Sensor
				sensorMarshalErr := json.Unmarshal([]byte(dbDeviceSensors.String), &sensorColl)
				if sensorMarshalErr == nil {
					tmpDevice.Sensors = sensorColl
				}
				if dbCalibrationValues.Valid {

					var calibVals CalibrationRules

					calibValMarshalErr := json.Unmarshal([]byte(dbCalibrationValues.String), &calibVals)
					if calibValMarshalErr == nil {
						tmpDevice.Calibration = calibVals
					}
				}
				if dbDeviceAttributes.Valid {
					attmap := make(map[string]string)
					attMarshalErr := json.Unmarshal([]byte(dbDeviceAttributes.String), &attmap)
					if attMarshalErr == nil {
						tmpDevice.Attributes = attmap
					}
				}
				if dbLat.Valid {
					tmpDevice.Lat = dbLat.Float64
				}
				if dbLng.Valid {
					tmpDevice.Lng = dbLng.Float64
				}
				if dbStatus.Valid {
					tmpDevice.LocationStatus = dbStatus.String
				}

				tmpDeviceMap[tmpDevice.MAC] = tmpDevice
			}

		}

	}
	d.Lock()
	//fmt.Println("temp dev map:	", tmpDeviceMap)
	d.List = tmpDeviceMap
	d.Unlock()
}

func (d *DeviceCollection) GetDevice(deviceMac string) Device {
	d.RLock()
	defer d.RUnlock()
	return d.List[deviceMac]
}
func (d *DeviceCollection) SetList(newList map[string]Device) {
	d.Lock()
	defer d.Unlock()
	d.List = newList

}

func (c CalibrationRules) GetCalibForPhenom(phenom string) CalibrationRule {

	for _, v := range c {
		if v.PhenomenonName == phenom {
			return v
		}

	}
	var blank CalibrationRule
	return blank
}
func basiccalibrationrules() {
	var tcalrules CalibrationRules
	var tcalrule CalibrationRule
	tcalrule.CalibrationFactors = make(map[string]float64)
	tcalrule.ModelName = "linearslopeandoffset"
	tcalrule.PhenomenonName = "Audio"
	tcalrule.CalibrationFactors["slope"] = 1
	tcalrule.CalibrationFactors["offset"] = 0
	tcalrules = append(tcalrules, tcalrule)
	var tcalrule1 CalibrationRule
	tcalrule1.CalibrationFactors = make(map[string]float64)

	tcalrule1.ModelName = "linearslopeandoffset"
	tcalrule1.PhenomenonName = "light"
	tcalrule1.CalibrationFactors["slope"] = 1
	tcalrule1.CalibrationFactors["offset"] = 0
	tcalrules = append(tcalrules, tcalrule1)
	var tcalrule2 CalibrationRule
	tcalrule2.CalibrationFactors = make(map[string]float64)

	tcalrule2.ModelName = "linearslopeandoffset"
	tcalrule2.PhenomenonName = "CO2"
	tcalrule2.CalibrationFactors["slope"] = 1
	tcalrule2.CalibrationFactors["offset"] = 0
	tcalrules = append(tcalrules, tcalrule2)

	tcalrulesJSON, err := json.MarshalIndent(tcalrules, " ", " ")
	if err == nil {
		fmt.Println(string(tcalrulesJSON))
	}

}
