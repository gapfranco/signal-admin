package clientstorage

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"time"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

// OpenTursoDB opens a libSQL connection to a remote Turso database.
func OpenTursoDB(primaryURL, authToken string) (*sql.DB, error) {
	u, err := url.Parse(primaryURL)
	if err != nil {
		return nil, fmt.Errorf("parse turso url: %w", err)
	}

	q := u.Query()
	q.Set("auth_token", authToken)
	u.RawQuery = q.Encode()

	db, err := sql.Open("libsql", u.String())
	if err != nil {
		return nil, fmt.Errorf("open remote: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping remote: %w", err)
	}

	return db, nil
}
