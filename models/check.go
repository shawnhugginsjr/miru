package models

import (
	"database/sql"
	"time"

	"github.com/besser/cron"
	"github.com/jmoiron/sqlx"
)

// Check represents a cron job for a URL.
type Check struct {
	ID          int64     `db:"id"`
	Name        string    `db:"name"`
	Cron        string    `db:"cron"`
	URL         string    `db:"url"`
	Status      string    `db:"status"`
	Active      bool      `db:"active"`
	Job         string    `db:"job"`
	LastContact time.Time `db:"last_contact"`
	NextContact time.Time `db:"next_contact"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

// NewCheck returns a pointer to a Check.
func NewCheck() *Check {
	now := time.Now()
	return &Check{
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// UpdateJob sets the new status along with updating the contact times.
func (c *Check) UpdateJob(status string) {
	if c.NextContact.IsZero() {
		c.NextContact = time.Now()
	}
	currentContact := c.NextContact
	c.LastContact = currentContact
	c.Status = status
	s, _ := cron.Parse(c.Cron)
	c.NextContact = s.Next(currentContact)
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
	job TEXT NOT NULL,
	last_contact DATETIME NOT NULL,
	next_contact DATETIME NOT NULL,
	created_at DATETIME NOT NULL,
	updated_at DATETIME NOT NULL
);
`
	getCheckByIDString   = `SELECT * FROM checks WHERE id = ? LIMIT 1;`
	updateCheckJobString = `UPDATE checks SET status = ?, last_contact = ?, next_contact = ?, updated_at = ? WHERE id = ?;`
	insertCheckString    = `INSERT INTO checks (name, cron, url, status, active, job, last_contact, next_contact, created_at, updated_at) VALUES (?, ?,?,?,?,?,?,?,?,?);`
)

// GetCheckByID gets a Check from an ID.
func GetCheckByID(db *sqlx.DB, id int64) (*Check, error) {
	var c Check
	err := db.Get(&c, getCheckByIDString, id)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// UpdateCheckJob updates a checks row with the new status and contact times.
func UpdateCheckJob(db *sqlx.DB, c *Check) (sql.Result, error) {
	c.UpdatedAt = time.Now()
	return db.Exec(updateCheckJobString, c.Status, c.LastContact, c.NextContact, c.UpdatedAt, c.ID)
}

// InsertCheck adds a Check to the database.
func InsertCheck(db *sqlx.DB, c *Check) (sql.Result, error) {
	return db.Exec(insertCheckString,
		c.Name,
		c.Cron,
		c.URL,
		c.Status,
		c.Active,
		c.Job,
		c.LastContact,
		c.NextContact,
		c.CreatedAt,
		c.UpdatedAt)
}
