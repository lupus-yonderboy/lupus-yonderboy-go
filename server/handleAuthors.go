package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/lupus-yonderboy/lupus-yonderboy-go/db"
	"github.com/lupus-yonderboy/lupus-yonderboy-go/types"
)

var Authors = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	var authors []types.Author
	var rows *sql.Rows

	switch r.Method {
	case "GET":
		query := `
        SELECT
          id,                       -- 1
          name,                     -- 2
          date_created,             -- 3
          date_updated,             -- 4
          bio,                      -- 5
          image                     -- 6
        FROM authors
      `

		var err error
		rows, err = db.DB.Query(query)
		if err != nil {
			panic(err)
		}
	}

	for rows.Next() {
		var Id uint                     // 1
		var Name string                 // 2
		var DateCreated time.Time       // 3
		var DateUpdated types.NullTime  // 4
		var Bio types.NullString        // 5
		var Image types.NullInt64       // 6

		rows.Scan(
      &Id,                          // 1
      &Name,                        // 2
      &DateCreated,                 // 3
      &DateUpdated,                 // 4
      &Bio,                         // 5
      &Image,                       // 6
    )

		authors = append(authors, types.Author{
			Id:          Id,              // 1
			Name:        Name,            // 2
			DateCreated: DateCreated,     // 3
			DateUpdated: DateUpdated,     // 4
			Bio:         Bio,             // 5
			Image:       Image,           // 6
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(authors)
})
