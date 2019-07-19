package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/pkg/errors"

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
	check.URL = "https://godoc.org/github.com/pkg/errors"
	check.Cron = "@hourly"
	result, err := models.InsertCheck(db, check)
	if err != nil {
		log.Fatal(err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		log.Fatal(err)
	}

	checkURL(db, id, check.URL)
}

// checkURL gets the status code of the url and sets the contact times.
func checkURL(db *sqlx.DB, checkID int64, url string) {
	check, err := models.GetCheckByID(db, checkID)
	if err != nil {
		log.Print(errors.Wrap(err, "Could not get Check model."))
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		errText := fmt.Sprintf("GET request to %s failed", url)
		log.Print(errors.Wrap(err, errText))
		return
	}
	defer resp.Body.Close()

	check.UpdateJob(resp.Status)
	_, err = models.UpdateCheckJob(db, check)
	if err != nil {
		log.Print(errors.Wrap(err, "checks row job could not be updated."))
	}
}
