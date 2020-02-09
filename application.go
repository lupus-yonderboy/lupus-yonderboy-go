package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/rs/cors"
)

// *****************************************************************************
// ****** TYPES ****************************************************************
// *****************************************************************************

type Post struct {
	Id          uint
	Title       string
	DateCreated time.Time
	DateUpdated time.Time
	ShortTitle  string
	Content     string
	Author      uint
	Image       uint
	Archived 		bool
}

type Author struct {
	Id          uint
	Name        string
	DateCreated time.Time
	DateUpdated time.Time
	Bio         string
	Image       uint
	Archived		bool
}

// *****************************************************************************
// ****** SCHEMA ***************************************************************
// *****************************************************************************

// CREATE TABLE posts (
//   id SERIAL PRIMARY KEY,
//   title VARCHAR NOT NULL,
//   date_created TIMESTAMPTZ NOT NULL,
//   date_updated TIMESTAMPTZ,
//   short_title VARCHAR,
//   content TEXT,
//   author INTEGER NOT NULL,
//   image INTEGER
// );

// ALTER TABLE posts
// ADD archived BOOL DEFAULT false;

// CREATE TABLE authors (
//   id SERIAL PRIMARY KEY,
//   name VARCHAR NOT NULL,
//   date_created TIMESTAMPTZ NOT NULL,
//   date_updated TIMESTAMPTZ,
//   bio VARCHAR,
//   image INTEGER
// );

// ALTER TABLE authors
// ADD archived BOOL DEFAULT false;

// CREATE TABLE images (
//   id SERIAL PRIMARY KEY,
//   title VARCHAR NOT NULL,
//   url VARCHAR NOT NULL,
//   date_created TIMESTAMPTZ NOT NULL,
//   date_updated TIMESTAMPTZ,
//   description VARCHAR,
//   link VARCHAR
// );

// ALTER TABLE images
// ADD archived BOOL DEFAULT false;

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
		AllowedOrigins:   []string{"https://lupus-yonderboy.github.io", "http://localhost:3000"},
		AllowCredentials: true,
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut},
		AllowedHeaders:   []string{"Token", "Host", "User-Agent", "Accept", "Content-Length", "Content-Type", "Origin"},
	})

	handler := c.Handler(mux)

	mux.Handle("/", root)
	mux.Handle("/posts/", Posts)
	mux.Handle("/authors/", Authors)
	mux.Handle("/"+os.Getenv("ONE_TIME"), OneTime)

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

func newNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}

var OneTime = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
  alterAuthorsTable := `
    alter table authors
		add archived bool default false;
  `

  alterPostsTable := `
    alter table posts
		add archived bool default false;
  `

  alterImagesTable := `
    alter table images
		add archived bool default false;
  `

  var err error
  _, err = DB.Query(alterAuthorsTable)
  if err != nil {
    panic(err)
  }

  _, err = DB.Query(alterPostsTable)
  if err != nil {
    panic(err)
  }

  _, err = DB.Query(alterImagesTable)
  if err != nil {
    panic(err)
  }
})

// *****************************************************************************
// ****** AUTHORS **************************************************************
// *****************************************************************************

var Authors = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var authors []Author
	var rows *sql.Rows

	requestPath := r.URL.Path
	pathSplit := strings.Split(requestPath, "/")
	paramAuthorId := pathSplit[2]

	switch r.Method {
	case "GET":
		query := `
        SELECT id,
               name,
               date_created,
               date_updated,
               bio,
               image,
							 archived
        FROM authors
				WHERE NOT archived
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
          name,
          date_created,
          date_updated,
          bio,
          image
        ) VALUES (
          $1,
          current_timestamp,
          current_timestamp,
          $2,
          $3
        ) RETURNING
          id,
          name,
          date_created,
          date_updated,
          bio,
          image,
					archived
      `

		rows, err = DB.Query(query,
			author.Name,
			author.Bio,
			author.Image,
		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	case "PUT":
		if paramAuthorId == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		author := &Author{}

		err := json.NewDecoder(r.Body).Decode(author)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query := `
				UPDATE authors
				SET
					name = COALESCE($1, name),
					date_updated = current_timestamp,
					bio = COALESCE($2, bio),
					image = CASE
						WHEN $3 = 0 THEN image
						ELSE $3
					END,
					archived = COALESCE($4, archived)
				WHERE id = $5 RETURNING
					id,
					name,
					date_created,
					date_updated,
					bio,
					image,
					archived
			`

		rows, err = DB.Query(query,
			newNullString(author.Name),
			newNullString(author.Bio),
			author.Image,
			author.Archived,
			paramAuthorId,
		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} // close switch

	for rows.Next() {
		var Id uint
		var Name string
		var DateCreated time.Time
		var DateUpdated time.Time
		var Bio string
		var Image uint
		var Archived bool

		rows.Scan(
			&Id,
			&Name,
			&DateCreated,
			&DateUpdated,
			&Bio,
			&Image,
			&Archived,
		)

		authors = append(authors, Author{
			Id:          Id,
			Name:        Name,
			DateCreated: DateCreated,
			DateUpdated: DateUpdated,
			Bio:         Bio,
			Image:       Image,
			Archived:		 Archived,
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

	requestPath := r.URL.Path
	pathSplit := strings.Split(requestPath, "/")
	paramPostId := pathSplit[2]

	switch r.Method {
	case "GET":
		query := `
        SELECT id,
               title,
               date_created,
               date_updated,
               short_title,
               content,
               author,
               image,
							 archived
        FROM posts
				WHERE NOT archived
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
          title,
          date_created,
          date_updated,
          short_title,
          content,
          author,
          image
        ) VALUES (
          $1,
          current_timestamp,
          current_timestamp,
          $2,
          $3,
          $4,
          $5
        ) RETURNING
          id,
          title,
          date_created,
          date_updated,
          short_title,
          content,
          author,
          image,
					archived
      `

		rows, err = DB.Query(query,
			post.Title,
			post.ShortTitle,
			post.Content,
			post.Author,
			post.Image,
		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

	case "PUT":
		if paramPostId == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		post := &Post{}

		err := json.NewDecoder(r.Body).Decode(post)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		query := `
				UPDATE posts
				SET
					title = COALESCE($1, title),
					date_updated = current_timestamp,
					short_title = COALESCE($2, short_title),
					content = COALESCE($3, content),
					author = CASE
						WHEN $4 = 0 THEN author
						ELSE $4
					END,
					image = CASE
						WHEN $5 = 0 THEN image
						ELSE $5
					END,
					archived = COALESCE($6, archived)
				WHERE id = $7 RETURNING
					id,
					title,
					date_created,
					date_updated,
					short_title,
					content,
					author,
					image,
					archived
			`

		rows, err = DB.Query(query,
			newNullString(post.Title),
			newNullString(post.ShortTitle),
			newNullString(post.Content),
			post.Author,
			post.Image,
			post.Archived,
			paramPostId,
		)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	} // close switch

	for rows.Next() {
		var Id uint
		var Title string
		var DateCreated time.Time
		var DateUpdated time.Time
		var ShortTitle string
		var Content string
		var Author uint
		var Image uint
		var Archived bool

		rows.Scan(
			&Id,
			&Title,
			&DateCreated,
			&DateUpdated,
			&ShortTitle,
			&Content,
			&Author,
			&Image,
			&Archived,
		)

		posts = append(posts, Post{
			Id:          Id,
			Title:       Title,
			DateCreated: DateCreated,
			DateUpdated: DateUpdated,
			ShortTitle:  ShortTitle,
			Content:     Content,
			Author:      Author,
			Image:       Image,
			Archived:		 Archived,
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
