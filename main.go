package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/pkg/errors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
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

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/", getAllCheckers)
	r.Route("/checkers", func(r chi.Router) {
		r.Get("/new", newChecker)        // GET /checkers/new
		r.Post("/", createChecker)       // POST /checkers
		r.Get("/{id}", getChecker)       // GET /checkers/10
		r.Delete("/{id}", deleteChecker) // DELETE /checkers/10
	})

	fmt.Println("Listening on port :8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

func getAllCheckers(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func newChecker(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/form.html"))
	tmpl.Execute(w, nil)
}

func createChecker(w http.ResponseWriter, r *http.Request) {
	c := models.CheckForm{}
	err := c.ExtractFormData(r)
	if err != nil {
		// TODO: Redirect not functional.
		http.Redirect(w, r, "/checkers/new", 200)
		return
	}

	w.Write([]byte("hi"))
}

func getChecker(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func deleteChecker(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
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
