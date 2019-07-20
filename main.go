package main

import (
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/besser/cron"
	"github.com/pkg/errors"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shawnhugginsjr/miru/models"
)

var cronRunner *cron.Cron
var db *sqlx.DB

func main() {
	var err error
	db, err = sqlx.Open("sqlite3", "./checks.db")
	if err != nil {
		log.Fatal(err)
	}

	cronRunner = cron.New()
	cronRunner.Start()

	db.Exec(models.CheckSchema)
	initCronJobs(db, cronRunner)

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
	checkOption := models.CheckOptions{}
	err := checkOption.ExtractFormData(r)
	if err != nil {
		// TODO: Redirect not functional.
		fmt.Println(err)
		http.Redirect(w, r, "/checkers/new", 200)
		return
	}

	check := models.NewCheckFromOptions(&checkOption)
	err = check.Insert(db, cronRunner)

	if err != nil {
		// log
		// send failure page
		fmt.Println(err)
		return
	}

	w.Write([]byte("succusful"))
}

func getChecker(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func deleteChecker(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("hi"))
}

func initCronJobs(db *sqlx.DB, cr *cron.Cron) error {
	var checks []models.Check
	err := db.Select(&checks, "SELECT * FROM checks LIMIT 10")
	if err != nil {
		return errors.Wrap(err, "Could not query database")
	}
	for _, c := range checks {
		fmt.Println("Starting check")
		if !c.Active {
			continue
		}

		err = c.RefreshNextContact(db)
		if err != nil {
			fmt.Println(err)
			continue
		}

		entryID, err := cr.AddFunc(c.Cron, c.CreateJobFunc(db))
		if err != nil {
			fmt.Print(err)
			continue
		}

		err = c.SetJob(db, entryID)
		if err != nil {
			fmt.Print(err)
			cronRunner.Remove(entryID)
			continue
		}
		fmt.Printf("Cron Job \"%s\" is running\n", c.Name)
	}
	return nil
}
