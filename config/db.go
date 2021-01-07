package config

import (
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
