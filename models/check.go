package models

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// Check represents a cron job for a URL.
type Check struct {
	ID          int       `db:"id"`
	Name        string    `db:"name"`
	Cron        string    `db:"cron"`
	URL         string    `db:"url"`
	Status      string    `db:"url"`
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
	last_contact INTEGER NOT NULL,
	next_contact INTEGER NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);
`
	insertCheckString = `INSERT INTO checks (name, cron, url, status, active, job, last_contact, next_contact, created_at, updated_at) VALUES (?, ?,?,?,?,?,?,?,?,?)`
)

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
