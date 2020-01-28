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

  "github.com/rs/cors"
	_ "github.com/lib/pq"
)

// types

type Post struct {
	Id          uint      // 1
	Title       string    // 2
	DateCreated time.Time // 3
  DateUpdated time.Time // 4
	ShortTitle  string    // 5
	Content     string    // 6
	Author      uint      // 7
	Image       uint      // 8
}

type Author struct {
	Id          uint
	Name        string
	DateCreated time.Time
  DateUpdated time.Time
	Bio         string
	Image       uint
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
               date_updated, -- 4
               bio,          -- 5
               image         -- 6
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
		var DateUpdated time.Time // 4
		var Bio string            // 5
		var Image uint            // 6

		rows.Scan(
			&Id,          // 1
			&Name,        // 2
			&DateCreated, // 3
      &DateUpdated, // 4
			&Bio,         // 5
			&Image,       // 6
		)

		authors = append(authors, Author{
			Id:          Id,          // 1
			Name:        Name,        // 2
			DateCreated: DateCreated, // 3
      DateUpdated: DateUpdated, // 4
			Bio:         Bio,         // 5
			Image:       Image,       // 6
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
               date_updated, -- 4
               short_title,  -- 5
               content,      -- 6
               author,       -- 7
               image         -- 8
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
          date_updated,       -- 4
          short_title,        -- 5
          content,            -- 6
          author,             -- 7
          image               -- 8
        ) VALUES (
                              -- 1
          $1,                 -- 2
          current_timestamp,  -- 3
          current_timestamp,  -- 4
          $2,                 -- 5
          $3,                 -- 6
          $4,                 -- 7
          $5                  -- 8
        ) RETURNING id,       -- 1
          title,              -- 2
          date_created,       -- 3
          date_updated,       -- 4
          short_title,        -- 5
          content,            -- 6
          author,             -- 7
          image               -- 8
      `

		rows, err = DB.Query(query,
			                  // 1
			post.Title,       // 2 -- $1
			                  // 3
                        // 4
			post.ShortTitle,  // 5 -- $2
			post.Content,     // 6 -- $3
			post.Author,      // 7 -- $4
			post.Image,       // 8 -- $5
		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
      fmt.Println(err)
			return
		}
	} // close switch

	for rows.Next() {
		var Id uint               // 1
		var Title string          // 2
		var DateCreated time.Time // 3
    var DateUpdated time.Time // 4
		var ShortTitle string     // 5
		var Content string        // 6
		var Author uint           // 7
		var Image uint            // 8

		rows.Scan(
			&Id,          // 1
			&Title,       // 2
			&DateCreated, // 3
      &DateUpdated, // 4
			&ShortTitle,  // 5
			&Content,     // 6
			&Author,      // 7
			&Image,       // 8
		)

		posts = append(posts, Post{
			Id:          Id,          // 1
			Title:       Title,       // 2
			DateCreated: DateCreated, // 3
      DateUpdated: DateUpdated, // 4
			ShortTitle:  ShortTitle,  // 5
			Content:     Content,     // 6
			Author:      Author,      // 7
			Image:       Image,       // 8
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
