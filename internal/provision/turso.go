package provision

import (
	"context"
	"errors"
	"fmt"
	"strings"

	turso "github.com/invertase/turso-api-client-go"
)

const defaultGroup = "default"

// TursoAPI wraps the Turso Platform API for database provisioning.
type TursoAPI struct {
	client *turso.Client
}

// NewTursoAPI creates a Turso Platform API client.
func NewTursoAPI(org, token string) (*TursoAPI, error) {
	client, err := turso.NewClient(turso.Config{
		Org:   org,
		Token: token,
	})
	if err != nil {
		return nil, fmt.Errorf("criar cliente Turso: %w", err)
	}
	return &TursoAPI{client: client}, nil
}

// DatabaseExists reports whether a database with the given name already exists.
func (t *TursoAPI) DatabaseExists(ctx context.Context, name string) (bool, error) {
	_, err := t.client.Databases.Get(ctx, name)
	if err == nil {
		return true, nil
	}

	var apiErr *turso.ClientError
	if errors.As(err, &apiErr) && apiErr.Status == 404 {
		return false, nil
	}
	return false, err
}

// CreateDatabase creates a new database in the default group.
func (t *TursoAPI) CreateDatabase(ctx context.Context, name string) (*turso.CreatedDatabase, error) {
	group := defaultGroup
	return t.client.Databases.Create(ctx, name, &turso.DatabaseCreateOptions{
		Group: &group,
	})
}

// CreateDatabaseToken issues a full-access token for the database.
func (t *TursoAPI) CreateDatabaseToken(ctx context.Context, name string) (string, error) {
	auth := "full-access"
	token, err := t.client.Databases.CreateToken(ctx, name, &turso.DatabaseCreateTokenOptions{
		Authorization: auth,
	})
	if err != nil {
		return "", err
	}
	return token.JWT, nil
}

// LibSQLURL builds a libsql connection URL from a Turso hostname.
func LibSQLURL(hostname string) string {
	hostname = strings.TrimSpace(hostname)
	if strings.HasPrefix(hostname, "libsql://") {
		return hostname
	}
	return "libsql://" + hostname
}
