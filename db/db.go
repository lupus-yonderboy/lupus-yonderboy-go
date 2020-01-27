package db

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

var DB *sql.DB

var (
	host     = os.Getenv("RDS_HOSTNAME")
	port, _  = strconv.ParseUint(os.Getenv("RDS_PORT"), 10, 64)
	dbname   = os.Getenv("RDS_DB_NAME")
  user     = os.Getenv("RDS_USERNAME")
  password = os.Getenv("RDS_PASSWORD")
)

func Connect() {
	info := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", host, port, dbname, user, password)

	var err error
	DB, err = sql.Open("postgres", info)
	if err != nil {
		panic(err)
	}

	// defer DB.Close()

	err = DB.Ping()
	if err != nil {
		panic(err)
	}
}
