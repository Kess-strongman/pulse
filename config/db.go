package config

import (
	"fmt"
	"log"

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
