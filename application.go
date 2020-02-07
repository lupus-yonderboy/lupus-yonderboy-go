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

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

// *****************************************************************************
// ****** TYPES ****************************************************************
// *****************************************************************************

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
	Id          uint      // 1
	Name        string    // 2
	DateCreated time.Time // 3
	DateUpdated time.Time // 4
	Bio         string    // 5
	Image       uint      // 6
}

// *****************************************************************************
// ****** SCHEMA ***************************************************************
// *****************************************************************************

// CREATE TABLE posts (
//   id SERIAL PRIMARY KEY,             -- 1
//   title VARCHAR NOT NULL,            -- 2
//   date_created TIMESTAMPTZ NOT NULL, -- 3
//   date_updated TIMESTAMPTZ,          -- 4
//   short_title VARCHAR,               -- 5
//   content TEXT,                      -- 6
//   author INTEGER NOT NULL,           -- 7
//   image INTEGER                      -- 8
// );

// CREATE TABLE authors (
//   id SERIAL PRIMARY KEY,             -- 1
//   name VARCHAR NOT NULL,             -- 2
//   date_created TIMESTAMPTZ NOT NULL, -- 3
//   date_updated TIMESTAMPTZ,          -- 4
//   bio VARCHAR,                       -- 5
//   image INTEGER                      -- 6
// );

// CREATE TABLE images (
//   id SERIAL PRIMARY KEY,             -- 1
//   title VARCHAR NOT NULL,            -- 2
//   url VARCHAR NOT NULL,              -- 3
//   date_created TIMESTAMPTZ NOT NULL, -- 4
//   date_updated TIMESTAMPTZ,          -- 5
//   description VARCHAR,               -- 6
//   link VARCHAR                       -- 7
// );

// *****************************************************************************
// ****** DATABASE *************************************************************
// *****************************************************************************

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

// *****************************************************************************
// ****** SERVER ***************************************************************
// *****************************************************************************

func Start() {
	mux := http.NewServeMux()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"https://lupus-yonderboy.github.io"},
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost},
		AllowedHeaders:   []string{"Token", "Host", "User-Agent", "Accept", "Content-Length", "Content-Type", "Origin"},
	})

	handler := c.Handler(mux)

	mux.Handle("/", root)
	mux.Handle("/posts", Posts)
	mux.Handle("/authors", Authors)

	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	cert := os.Getenv("CERT")
	privKey := os.Getenv("PRIV_KEY")

	log.Fatal(http.ListenAndServeTLS(":"+port, cert, privKey, handler))
}

var root = http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode("Hi.")
})

// *****************************************************************************
// ****** AUTHORS **************************************************************
// *****************************************************************************

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

	case "POST":
		author := &Author{}

		err := json.NewDecoder(r.Body).Decode(author)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query := `
        INSERT INTO authors (
                             -- 1
          name,              -- 2
          date_created,      -- 3
          date_updated,      -- 4
          bio,               -- 5
          image              -- 6
        ) VALUES (
                             -- 1
          $1,                -- 2
          current_timestamp, -- 3
          current_timestamp, -- 4
          $2,                -- 5
          $3                 -- 6
        ) RETURNING
          id,                -- 1
          name,              -- 2
          date_created,      -- 3
          date_updated,      -- 4
          bio,               -- 5
          image              -- 6
      `

		rows, err = DB.Query(query,
              		  // 1
			author.Name,  // 2 -- $1
              		  // 3
              		  // 4
			author.Bio,   // 5 -- $2
			author.Image, // 6 -- $3
		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} // close switch

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

// *****************************************************************************
// ****** POSTS ****************************************************************
// *****************************************************************************

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
          title,             -- 2
          date_created,      -- 3
          date_updated,      -- 4
          short_title,       -- 5
          content,           -- 6
          author,            -- 7
          image              -- 8
        ) VALUES (
                             -- 1
          $1,                -- 2
          current_timestamp, -- 3
          current_timestamp, -- 4
          $2,                -- 5
          $3,                -- 6
          $4,                -- 7
          $5                 -- 8
        ) RETURNING
          id,                -- 1
          title,             -- 2
          date_created,      -- 3
          date_updated,      -- 4
          short_title,       -- 5
          content,           -- 6
          author,            -- 7
          image              -- 8
      `

		rows, err = DB.Query(query,
			                 // 1
			post.Title,      // 2 -- $1
                  		 // 3
                  		 // 4
			post.ShortTitle, // 5 -- $2
			post.Content,    // 6 -- $3
			post.Author,     // 7 -- $4
			post.Image,      // 8 -- $5
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

// *****************************************************************************
// ****** MAIN *****************************************************************
// *****************************************************************************

func main() {
	Connect()
	Start()
}
