package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/lib/pq"
	"github.com/rs/cors"
)

// types

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
	ShortTitle  NullString
	Content     NullString
	Author      uint
	Image       NullInt64
	DateUpdated NullTime
}

type Author struct {
	Id          uint
	Name        string
	DateCreated time.Time
	Bio         NullString
	Image       NullInt64
	DateUpdated NullTime
}

// database

var DB *sql.DB

var (
	host      = os.Getenv("RDS_HOSTNAME")
	dbPort, _ = strconv.ParseUint(os.Getenv("RDS_PORT"), 10, 64)
	dbname    = os.Getenv("RDS_DB_NAME")
	user      = os.Getenv("RDS_USERNAME")
	password  = os.Getenv("RDS_PASSWORD")
)

func Connect() {
	info := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=disable", host, dbPort, dbname, user, password)

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

// server

func Start() {
	origin := "https://lupus-yonderboy.github.io/lupus-yonderboy"

	mux := http.NewServeMux()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{origin},
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet},
		AllowedHeaders:   []string{"Token", "Host", "User-Agent", "Accept", "Content-Length", "Content-Type"},
	})

	handler := c.Handler(mux)

	mux.Handle("/", root)
	mux.Handle("/posts", Posts)
	mux.Handle("/authors", Authors)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	log.Fatal(http.ListenAndServe(":"+port, handler))
}

var root = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Hi.")
})

var Authors = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var authors []Author
	var rows *sql.Rows

	switch r.Method {
	case "GET":
		query := `
        SELECT id,           -- 1
               name,         -- 2
               date_created, -- 3
               bio,          -- 4
               image         -- 5
               date_updated, -- 6
        FROM authors
      `

		var err error
		rows, err = DB.Query(query)
		if err != nil {
			panic(err)
		}
	}

	for rows.Next() {
		var Id uint               // 1
		var Name string           // 2
		var DateCreated time.Time // 3
		var Bio NullString        // 4
		var Image NullInt64       // 5
		var DateUpdated NullTime  // 6

		rows.Scan(
			&Id,          // 1
			&Name,        // 2
			&DateCreated, // 3
			&Bio,         // 4
			&Image,       // 5
			&DateUpdated, // 6
		)

		authors = append(authors, Author{
			Id:          Id,          // 1
			Name:        Name,        // 2
			DateCreated: DateCreated, // 3
			Bio:         Bio,         // 4
			Image:       Image,       // 5
			DateUpdated: DateUpdated, // 6
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
})

var Posts = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var posts []Post
	var rows *sql.Rows

	switch r.Method {
	case "GET":
		query := `
        SELECT id,           -- 1
               title,        -- 2
               date_created, -- 3
               short_title,  -- 4
               content,      -- 5
               author,       -- 6
               image,        -- 7
               date_updated  -- 8
        FROM posts
      `

		var err error
		rows, err = DB.Query(query)
		if err != nil {
			panic(err)
		}

	case "POST":
		post := &Post{}

		err := json.NewDecoder(r.Body).Decode(post)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query := `
        INSERT INTO posts (
                              -- 1
          title,              -- 2
          date_created,       -- 3
          short_title,        -- 4
          content,            -- 5
          author,             -- 6
          image               -- 7
                              -- 8
        ) VALUES (
                              -- 1
          $1,                 -- 2
          $2,                 -- 3
          now(),              -- 4
          COALESCE($3, NULL), -- 5
          COALESCE($4, NULL), -- 6
          $5,                 -- 7
          $6                  -- 8
        ) RETURNING id,       -- 1
          title,              -- 2
          date_created,       -- 3
          short_title,        -- 4
          content,            -- 5
          author,             -- 6
          image,              -- 7
          date_updated        -- 8
      `

		rows, err = DB.Query(query,
			                 // 1
			post.Title,      // 2
			post.ShortTitle, // 3
			                 // 4
			post.Content,    // 5
			post.Author,     // 6
			post.Image,      // 7

		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} // close switch

	for rows.Next() {
		var Id uint               // 1
		var Title string          // 2
		var DateCreated time.Time // 3
		var ShortTitle NullString // 4
		var Content NullString    // 5
		var Author uint           // 6
		var Image NullInt64       // 7
		var DateUpdated NullTime  // 8

		rows.Scan(
			&Id,          // 1
			&Title,       // 2
			&DateCreated, // 3
			&ShortTitle,  // 4
			&Content,     // 5
			&Author,      // 6
			&Image,       // 7
			&DateUpdated, // 8
		)

		posts = append(posts, Post{
			Id:          Id,          // 1
			Title:       Title,       // 2
			DateCreated: DateCreated, // 3
			ShortTitle:  ShortTitle,  // 4
			Content:     Content,     // 5
			Author:      Author,      // 6
			Image:       Image,       // 7
			DateUpdated: DateUpdated, // 8
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
})

// main

func main() {
	Connect()
	Start()
}
