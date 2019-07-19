package models

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/besser/cron"
)

// CheckOptions holds the form post data for creating a Check.
type CheckOptions struct {
	Name   string
	Cron   string
	URL    string
	Active bool
}

// ExtractFormData extracts the form data into the CheckForm.
// An error is returned with the reason if the data is invalid.
func (c *CheckOptions) ExtractFormData(r *http.Request) error {
	c.Name = r.FormValue("name")
	c.Cron = r.FormValue("cron")
	c.URL = r.FormValue("url")
	if r.FormValue("active") == "true" {
		c.Active = true
	} else {
		c.Active = false
	}

	if len(c.Name) == 0 || len(c.Cron) == 0 || len(c.URL) == 0 {
		return errors.New("All fields need to be supplied")
	}

	_, err := cron.Parse(c.Cron)
	if err != nil {
		return fmt.Errorf("Invalid Cron Job: %s", err)
	}

	return nil
}
