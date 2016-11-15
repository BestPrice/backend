package sql

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// Client represents client to the sql database
type Client struct {
	db *sql.DB

	// Path to postgres database
	Path string
}

func (c *Client) Open() error {
	db, err := sql.Open("postgres", c.Path)
	if err != nil {
		return err
	}
	c.db = db
	_, err = db.Query("CREATE EXTENSION IF NOT EXISTS unaccent")
	return err
}

func (c *Client) Service() *Service {
	return &Service{db: c.db}
}
