package main

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shawnhugginsjr/miru/models"
)

func main() {
	db, err := sqlx.Open("sqlite3", "./checks.db")
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(models.CheckSchema)

	check := models.NewCheck()
	result, err := models.InsertCheck(db, check)
	if err != nil {
		log.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("id is: %d\n", id)
}
