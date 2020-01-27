package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lupus-yonderboy/lupus-yonderboy-go/db"
	"github.com/lupus-yonderboy/lupus-yonderboy-go/types"
)

var Posts = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var posts []types.Post
	var rows *sql.Rows

	switch r.Method {
	case "GET":
		query := `
        SELECT id,                  -- 1
               title,               -- 2
               date_created,        -- 3
               date_updated,        -- 4
               short_title,         -- 5
               content,             -- 6
               author,              -- 7
               image                -- 8
        FROM posts
      `

		var err error
		rows, err = db.DB.Query(query)
		if err != nil {
			panic(err)
		}
	}

	for rows.Next() {
		var Id uint                     // 1
		var Title string                // 2
		var DateCreated time.Time       // 3
		var DateUpdated types.NullTime  // 4
		var ShortTitle types.NullString // 5
		var Content types.NullString    // 6
		var Author uint                 // 7
		var Image types.NullInt64       // 8

		rows.Scan(
      &Id,                          // 1
      &Title,                       // 2
      &DateCreated,                 // 3
      &DateUpdated,                 // 4
      &ShortTitle,                  // 5
      &Content,                     // 6
      &Author,                      // 7
      &Image,                       // 8
    )

		posts = append(posts, types.Post{
			Id:          Id,               // 1
			Title:       Title,            // 2
			DateCreated: DateCreated,      // 3
			DateUpdated: DateUpdated,      // 4
			ShortTitle:  ShortTitle,       // 5
			Content:     Content,          // 6
			Author:      Author,           // 7
			Image:       Image,            // 8
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(posts)
})
