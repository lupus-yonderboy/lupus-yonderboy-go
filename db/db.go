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
	host    = os.Getenv("HOST")
	port, _ = strconv.ParseUint(os.Getenv("DB_PORT"), 10, 64)
	dbname  = os.Getenv("DB_NAME")
)

func Connect() {
	info := fmt.Sprintf("host=%s port=%d dbname=%s sslmode=disable", host, port, dbname)

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
