package main

import (
	"database/sql"
	"log"

	"github.com/adhocteam/soapbox/api"
)

func main() {
	db, err := sql.Open("postgres", "")
	if err != nil {
		log.Fatalf("opening db conn: %v", err)
	}

	s := api.NewServer(db, nil)
	_ = s
}
