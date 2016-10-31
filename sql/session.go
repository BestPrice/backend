package sql

import (
	"database/sql"

	"github.com/BestPrice/backend/bp"
)

// Session represents an authenticable connection to the database.
type Session struct {
	db *sql.DB

	// Authentication
	authenticator bp.Authenticator
	authToken     string
	// user          *bp.User

	// Services
	service Service
}

func newSession(db *sql.DB) *Session {
	s := &Session{db: db}
	s.service.session = s
	return s
}

func (s *Session) SetAuthToken(token string) {
	s.authToken = token
}

func (s *Session) Authenticate() error {
	return s.authenticator.Authenticate(s.authToken)
}

func (s *Session) Service() Service {
	return s.service
}
