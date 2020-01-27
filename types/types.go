package types

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type NullTime struct {
	pq.NullTime
}

type NullString struct {
	sql.NullString
}

type NullInt64 struct {
	sql.NullInt64
}

type Post struct {
	Id          uint
	Title       string
	DateCreated time.Time
	DateUpdated NullTime
	ShortTitle  NullString
	Content     NullString
	Author      uint
	Image       NullInt64
}

type Author struct {
	Id          uint
	Name        string
	DateCreated time.Time
	DateUpdated NullTime
	Bio         NullString
	Image       NullInt64
}
