package sql

import (
	"database/sql"

	"github.com/BestPrice/backend/bp"
	_ "github.com/lib/pq"
)

// Client represents client to the sql database
type Client struct {
	db *sql.DB

	// Path to postgres database
	Path string

	// Autenticator to use
	Authenticator bp.Authenticator
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

func (c *Client) Connect() *Session {
	s := newSession(c.db)
	s.authenticator = c.Authenticator
	return s
}
