package models

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/besser/cron"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// Check represents a cron job for a URL.
type Check struct {
	ID          int64        `db:"id"`
	Name        string       `db:"name"`
	Cron        string       `db:"cron"`
	URL         string       `db:"url"`
	Status      string       `db:"status"`
	Active      bool         `db:"active"`
	Job         cron.EntryID `db:"job"`
	LastContact time.Time    `db:"last_contact"`
	NextContact time.Time    `db:"next_contact"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   time.Time    `db:"updated_at"`
}

// --- SQL INTERACTIONS ---

// CheckSchema is the table format of Check
const (
	CheckSchema = `CREATE TABLE IF NOT EXISTS checks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NON NULL,
	cron TEXT NOT NULL,
	url TEXT NOT NULL,
	status TEXT NOT NULL,
	active INTEGER NOT NULL,
	job INTEGER NOT NULL,
	last_contact DATETIME NOT NULL,
	next_contact DATETIME NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
`
	getCheckByIDString    = `SELECT * FROM checks WHERE id = ? LIMIT 1;`
	saveJobStatusString   = `UPDATE checks SET status = ?, last_contact = ?, next_contact = ?, updated_at = ? WHERE id = ?;`
	updateCronJobIDString = `UPDATE checks SET job = ? WHERE id = ?`
	insertCheckString     = `INSERT INTO checks (name, cron, url, status, active, job, last_contact, next_contact, created_at, updated_at) VALUES (?, ?,?,?,?,?,?,?,?,?);`
)

// NewCheck returns a pointer to a Check.
func NewCheck() *Check {
	now := time.Now()
	return &Check{
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// NewCheckFromOptions returns a new Check using data from a CheckForm.
// It's expected that the CheckForm is already validated.
func NewCheckFromOptions(cf *CheckOptions) *Check {
	c := NewCheck()
	c.Name = cf.Name
	c.Cron = cf.Cron
	c.URL = cf.URL
	c.Active = cf.Active
	return c
}

// GetCheckByID gets a Check from an ID.
func GetCheckByID(db *sqlx.DB, id int64) (*Check, error) {
	var c Check
	err := db.Get(&c, getCheckByIDString, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// SetJobStatus sets the new status along with updating the contact times.
// These changes are then persisted in the database.
func (c *Check) SetJobStatus(db *sqlx.DB, status string) (sql.Result, error) {
	if c.NextContact.IsZero() {
		c.NextContact = time.Now()
	}
	currentContact := c.NextContact
	c.LastContact = currentContact
	c.Status = status
	s, _ := cron.Parse(c.Cron)
	c.NextContact = s.Next(currentContact)

	c.UpdatedAt = time.Now()
	return db.Exec(saveJobStatusString, c.Status, c.LastContact, c.NextContact, c.UpdatedAt, c.ID)
}

func (c *Check) SetJob(db *sqlx.DB, job cron.EntryID) error {
	c.Job = job
	_, err := db.Exec(updateCronJobIDString, c.Job, c.ID)
	if err != nil {
		return errors.Wrap(err, "Could not set Check to a new Job")
	}
	return nil
}

// CreateJobFunc returns a function that will hit a HTTP endpoint based
// from the receiver.
func (c *Check) CreateJobFunc(db *sqlx.DB) func() {
	return func() {
		check, err := GetCheckByID(db, c.ID)
		if err != nil {
			log.Print(errors.Wrap(err, "Could not get Check model."))
			return
		}

		resp, err := http.Get(c.URL)
		if err != nil {
			errText := fmt.Sprintf("GET request to %s failed", check.URL)
			log.Print(errors.Wrap(err, errText))
			return
		}
		defer resp.Body.Close()
		log.Printf("%s Check %d: %s %s", time.Now().Format("2006-01-02 15:04:05"),
			check.ID, check.URL, resp.Status)

		_, err = check.SetJobStatus(db, resp.Status)
		if err != nil {
			log.Print(errors.Wrap(err, "checks row job could not be updated."))
		}
	}
}

// Insert adds the Check into the database along with starting
// a cron job if the Check is active.
func (c *Check) Insert(db *sqlx.DB, cr *cron.Cron) error {
	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "Could not begin transaction")
	}

	result, err := tx.Exec(insertCheckString,
		c.Name, c.Cron, c.URL, c.Status, c.Active, c.Job,
		c.LastContact, c.NextContact, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "Could not insert Check")
	}

	id, err := result.LastInsertId()
	if err != nil {
		tx.Rollback()
		return errors.Wrap(err, "Could not get Check ID")
	}
	c.ID = id

	if c.Active {
		entryID, err := cr.AddFunc(c.Cron, c.CreateJobFunc(db))
		if err != nil {
			tx.Rollback()
			return errors.Wrap(err, "Adding Cron Job failed")
		}

		c.Job = entryID
		_, err = tx.Exec(updateCronJobIDString, c.Job, c.ID)
		if err != nil {
			cr.Remove(entryID)
			tx.Rollback()
			return errors.Wrap(err, "Could not set Cron Job ID to Check")
		}
	}
	tx.Commit()

	return nil
}
