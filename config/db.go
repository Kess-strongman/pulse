package config

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	pgx "github.com/jackc/pgx/v4/pgxpool"
)

//DB connection to database
var DB *pgx.Pool

//DBerr database error
var DBerr error

//InitDatabase to create connection
func InitDatabase(connectionString string) {

	//fmt.Println(connectionString)

	DB, DBerr = pgxpool.Connect(context.Background(), connectionString)

	log.Println("DB Connected")
	if DBerr != nil {
		fmt.Println("Unabled to Create DB Connection", DBerr)
	}
}

type DBrow struct {
	TS         time.Time
	LocationID string
	DeviceID   string
	Lat        float64
	Lng        float64
	Data       DBdatapayload
}

type DBdatapayload struct {
	Dstring map[string]string
	Dnum    map[string]float64
	Dbool   map[string]bool
}

func InsertTimeSeriesRows(newrows []DBrow) error {
	fmt.Println(newrows)
	insertSQL := "INSERT INTO pulsetsd (ts, location_id, device_id,lat, lng, data) VALUES " //ON CONFLICT ON CONSTRAINT tsd_pkey DO NOTHING;
	var holder []interface{}
	var parametercounter int64
	parametercounter = 1
	for _, tempro := range newrows {
		p1 := strconv.FormatInt(parametercounter, 10)
		p2 := strconv.FormatInt(parametercounter+1, 10)
		p3 := strconv.FormatInt(parametercounter+2, 10)
		p4 := strconv.FormatInt(parametercounter+3, 10)
		p5 := strconv.FormatInt(parametercounter+4, 10)
		p6 := strconv.FormatInt(parametercounter+5, 10)
		insertSQL = insertSQL + "($" + p1 + ",$" + p2 + ",$" + p3 + ",$" + p4 + ",$" + p5 + ",$" + p6 + "),"
		parametercounter = parametercounter + 6
		dataInBytes, DataMarshalErr := json.Marshal(tempro.Data)
		if DataMarshalErr != nil {
			return DataMarshalErr
		}
		holder = append(holder, tempro.TS, tempro.LocationID, tempro.DeviceID, tempro.Lat, tempro.Lng, string(dataInBytes))
	}
	insertSQL = strings.TrimRight(insertSQL, ",")
	insertSQL = insertSQL + "ON CONFLICT ON CONSTRAINT pulsetsd_pkey DO NOTHING;"
	fmt.Println(insertSQL)
	d, result := DB.Exec(context.Background(), insertSQL, holder...)
	fmt.Println(newrows, d.RowsAffected())
	if result != nil {
		fmt.Println("Error in create pulse TSD", result.Error())
		return result
	}
	return nil
}

func InsertTimeSeriesRow(newrow DBrow) error {
	var DBrows []DBrow
	DBrows = append(DBrows, newrow)
	return InsertTimeSeriesRows(DBrows)
}
func GetMessageForAppBetweenTimes(appID string, startTime time.Time, endTime time.Time) ([]DBrow, error) {
	var rowsToReturn []DBrow
	latestEntriesSQL := `select pt.location_id, pt.ts, pt.device_id,pt.lat, pt.lng,pt.data from pulsetsd pt
	where location_id = $1  and pt.ts > $2 and pt.ts < $3`

	fmt.Println("GetMessagesforAppBetween", startTime, " and ", endTime)

	rows, err := DB.Query(context.Background(), latestEntriesSQL, appID, startTime, endTime)

	if err != nil {
		fmt.Println(err)
		return rowsToReturn, err

	}
	defer rows.Close()
	var dbTS sql.NullString
	var dbLocID sql.NullString
	var dbDeviceID sql.NullString
	var dbLat sql.NullFloat64
	var dbLng sql.NullFloat64
	var dbData sql.NullString

	for rows.Next() {
		fmt.Println("Scanning")
		if scanErr := rows.Scan(&dbLocID, &dbTS, &dbDeviceID, &dbLat, &dbLng, &dbData); err != nil {
			return nil, scanErr
		}
		if dbLocID.Valid && dbTS.Valid && dbDeviceID.Valid && dbLat.Valid && dbLng.Valid && dbData.Valid {
			var newRow DBrow
			/*
				var newPayload DBdatapayload
				newPayload.Dbool = make(map[string]bool)
				newPayload.Dnum = make(map[string]float64)
				newPayload.Dstring = make(map[string]string)
			*/
			var dbPL DBdatapayload
			unmarshallerr := json.Unmarshal([]byte(dbData.String), &dbPL)
			if unmarshallerr != nil {
				return rowsToReturn, unmarshallerr
			}
			newRow.Data = dbPL
			newRow.DeviceID = dbDeviceID.String
			newRow.LocationID = dbLocID.String
			newRow.Lat = dbLat.Float64
			newRow.Lng = dbLng.Float64
			// time comes in as 2021-01-07 13:56:25.022257
			newTime, timeParseError := time.Parse("2006-01-02T15:04:05.000000Z", dbTS.String)
			if timeParseError != nil {
				return rowsToReturn, timeParseError

			}
			newRow.TS = newTime

			rowsToReturn = append(rowsToReturn, newRow)

		}

	}

	return rowsToReturn, nil

}

func GetLatestMessageForApp(appID string) ([]DBrow, error) {
	var rowsToReturn []DBrow
	latestEntriesSQL := `select sq.location_id, sq.ts, p2.device_id,p2.lat, p2.lng,p2.data from 
	(select p1.location_id , max(p1.ts) as ts from pulsetsd p1 where location_id = $1 group by p1.location_id) as sq 
	left join pulsetsd p2 on (p2.location_id = sq.location_id and p2.ts = sq.ts)`

	rows, err := DB.Query(context.Background(), latestEntriesSQL, appID)
	if err != nil {
		return rowsToReturn, err
	}
	defer rows.Close()
	var dbTS sql.NullString
	var dbLocID sql.NullString
	var dbDeviceID sql.NullString
	var dbLat sql.NullFloat64
	var dbLng sql.NullFloat64
	var dbData sql.NullString

	for rows.Next() {
		if scanErr := rows.Scan(&dbLocID, &dbTS, &dbDeviceID, &dbLat, &dbLng, &dbData); err != nil {
			return nil, scanErr
		}
		if dbLocID.Valid && dbTS.Valid && dbDeviceID.Valid && dbLat.Valid && dbLng.Valid && dbData.Valid {
			var newRow DBrow
			/*
				var newPayload DBdatapayload
				newPayload.Dbool = make(map[string]bool)
				newPayload.Dnum = make(map[string]float64)
				newPayload.Dstring = make(map[string]string)
			*/
			var dbPL DBdatapayload
			unmarshallerr := json.Unmarshal([]byte(dbData.String), &dbPL)
			if unmarshallerr != nil {
				return rowsToReturn, unmarshallerr
			}
			newRow.Data = dbPL
			newRow.DeviceID = dbDeviceID.String
			newRow.LocationID = dbLocID.String
			newRow.Lat = dbLat.Float64
			newRow.Lng = dbLng.Float64
			// time comes in as 2021-01-07 13:56:25.022257
			newTime, timeParseError := time.Parse("2006-01-02T15:04:05.000000Z", dbTS.String)
			if timeParseError != nil {
				return rowsToReturn, timeParseError

			}
			newRow.TS = newTime

			rowsToReturn = append(rowsToReturn, newRow)

		}

	}

	return rowsToReturn, nil

}

func GetLatestServiceStatusMessages() ([]DBrow, error) {
	var rowsToReturn []DBrow
	latestEntriesSQL := `select sq.location_id, sq.ts, p2.device_id,p2.lat, p2.lng,p2.data from 
	(select p1.location_id , max(p1.ts) as ts from pulsetsd p1 where device_id = 'ServiceStatusMessage' group by p1.location_id) as sq 
	left join pulsetsd p2 on (p2.location_id = sq.location_id and p2.ts = sq.ts)`

	rows, err := DB.Query(context.Background(), latestEntriesSQL)
	if err != nil {
		return rowsToReturn, err

	}
	defer rows.Close()
	var dbTS sql.NullString
	var dbLocID sql.NullString
	var dbDeviceID sql.NullString
	var dbLat sql.NullFloat64
	var dbLng sql.NullFloat64
	var dbData sql.NullString

	for rows.Next() {
		if scanErr := rows.Scan(&dbLocID, &dbTS, &dbDeviceID, &dbLat, &dbLng, &dbData); err != nil {
			return nil, scanErr
		}
		if dbLocID.Valid && dbTS.Valid && dbDeviceID.Valid && dbLat.Valid && dbLng.Valid && dbData.Valid {
			var newRow DBrow
			/*
				var newPayload DBdatapayload
				newPayload.Dbool = make(map[string]bool)
				newPayload.Dnum = make(map[string]float64)
				newPayload.Dstring = make(map[string]string)
			*/
			var dbPL DBdatapayload
			unmarshallerr := json.Unmarshal([]byte(dbData.String), &dbPL)
			if unmarshallerr != nil {
				return rowsToReturn, unmarshallerr
			}
			newRow.Data = dbPL
			newRow.DeviceID = dbDeviceID.String
			newRow.LocationID = dbLocID.String
			newRow.Lat = dbLat.Float64
			newRow.Lng = dbLng.Float64
			// time comes in as 2021-01-07 13:56:25.022257
			newTime, timeParseError := time.Parse("2006-01-02T15:04:05.000000Z", dbTS.String)
			if timeParseError != nil {
				return rowsToReturn, timeParseError

			}
			newRow.TS = newTime

			rowsToReturn = append(rowsToReturn, newRow)

		}

	}

	return rowsToReturn, nil

}

func GetLatestHelloMessages() ([]DBrow, error) {
	var rowsToReturn []DBrow
	latestEntriesSQL := `select sq.location_id, sq.ts, p2.device_id,p2.lat, p2.lng,p2.data from 
	(select p1.location_id , max(p1.ts) as ts from pulsetsd p1 where device_id = 'HelloMessage' group by p1.location_id) as sq 
	left join pulsetsd p2 on (p2.location_id = sq.location_id and p2.ts = sq.ts)`

	rows, err := DB.Query(context.Background(), latestEntriesSQL)
	if err != nil {
		return rowsToReturn, err

	}
	defer rows.Close()
	var dbTS sql.NullString
	var dbLocID sql.NullString
	var dbDeviceID sql.NullString
	var dbLat sql.NullFloat64
	var dbLng sql.NullFloat64
	var dbData sql.NullString

	for rows.Next() {
		if scanErr := rows.Scan(&dbLocID, &dbTS, &dbDeviceID, &dbLat, &dbLng, &dbData); err != nil {
			return nil, scanErr
		}
		if dbLocID.Valid && dbTS.Valid && dbDeviceID.Valid && dbLat.Valid && dbLng.Valid && dbData.Valid {
			var newRow DBrow
			/*
				var newPayload DBdatapayload
				newPayload.Dbool = make(map[string]bool)
				newPayload.Dnum = make(map[string]float64)
				newPayload.Dstring = make(map[string]string)
			*/
			var dbPL DBdatapayload
			unmarshallerr := json.Unmarshal([]byte(dbData.String), &dbPL)
			if unmarshallerr != nil {
				return rowsToReturn, unmarshallerr
			}
			newRow.Data = dbPL
			newRow.DeviceID = dbDeviceID.String
			newRow.LocationID = dbLocID.String
			newRow.Lat = dbLat.Float64
			newRow.Lng = dbLng.Float64
			// time comes in as 2021-01-07 13:56:25.022257
			newTime, timeParseError := time.Parse("2006-01-02T15:04:05.000000Z", dbTS.String)
			if timeParseError != nil {
				return rowsToReturn, timeParseError

			}
			newRow.TS = newTime

			rowsToReturn = append(rowsToReturn, newRow)

		}

	}

	return rowsToReturn, nil

}

func GetLatestEntries() ([]DBrow, error) {
	var rowsToReturn []DBrow
	latestEntriesSQL := `select sq.location_id, sq.ts, p2.device_id,p2.lat, p2.lng,p2.data from 
	(select p1.location_id , max(p1.ts) as ts from pulsetsd p1 group by p1.location_id) as sq 
	left join pulsetsd p2 on (p2.location_id = sq.location_id and p2.ts = sq.ts)`

	rows, err := DB.Query(context.Background(), latestEntriesSQL)
	if err != nil {
		return rowsToReturn, err

	}
	defer rows.Close()
	var dbTS sql.NullString
	var dbLocID sql.NullString
	var dbDeviceID sql.NullString
	var dbLat sql.NullFloat64
	var dbLng sql.NullFloat64
	var dbData sql.NullString

	for rows.Next() {
		if scanErr := rows.Scan(&dbLocID, &dbTS, &dbDeviceID, &dbLat, &dbLng, &dbData); err != nil {
			return nil, scanErr
		}
		if dbLocID.Valid && dbTS.Valid && dbDeviceID.Valid && dbLat.Valid && dbLng.Valid && dbData.Valid {
			var newRow DBrow
			/*
				var newPayload DBdatapayload
				newPayload.Dbool = make(map[string]bool)
				newPayload.Dnum = make(map[string]float64)
				newPayload.Dstring = make(map[string]string)
			*/
			var dbPL DBdatapayload
			unmarshallerr := json.Unmarshal([]byte(dbData.String), &dbPL)
			if unmarshallerr != nil {
				return rowsToReturn, unmarshallerr
			}
			newRow.Data = dbPL
			newRow.DeviceID = dbDeviceID.String
			newRow.LocationID = dbLocID.String
			newRow.Lat = dbLat.Float64
			newRow.Lng = dbLng.Float64
			// time comes in as 2021-01-07 13:56:25.022257
			newTime, timeParseError := time.Parse("2006-01-02T15:04:05.000000Z", dbTS.String)
			if timeParseError != nil {
				return rowsToReturn, timeParseError

			}
			newRow.TS = newTime

			rowsToReturn = append(rowsToReturn, newRow)

		}

	}

	return rowsToReturn, nil

}
