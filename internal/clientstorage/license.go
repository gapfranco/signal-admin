package clientstorage

import (
	"context"
	"database/sql"
	"strings"
)

type License struct {
	Nome     string
	ClientID string
	Validade string
	Status   string
}

// InsertLicense inserts a license row into the remote database.
func InsertLicense(ctx context.Context, db *sql.DB, lic License) error {
	var validade any
	if strings.TrimSpace(lic.Validade) != "" {
		validade = lic.Validade
	}

	status := "active"
	if trimmedStatus := strings.TrimSpace(lic.Status); trimmedStatus != "" {
		status = trimmedStatus
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO license (nome, client_id, validade, status) VALUES (?, ?, ?, ?)`,
		lic.Nome, lic.ClientID, validade, status,
	)
	return err
}
