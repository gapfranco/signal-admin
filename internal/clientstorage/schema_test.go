package clientstorage_test

import (
	"context"
	"database/sql"
	"testing"

	"signal-admin/internal/clientstorage"

	_ "modernc.org/sqlite"
)

func TestInitRemoteSchemaCreatesTablesAndIsIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	ctx := context.Background()
	if err := clientstorage.InitRemoteSchema(ctx, db); err != nil {
		t.Fatalf("InitRemoteSchema() error = %v", err)
	}
	if err := clientstorage.InitRemoteSchema(ctx, db); err != nil {
		t.Fatalf("second InitRemoteSchema() error = %v", err)
	}

	for _, table := range []string{"event", "cluster_lock", "remote_metadata", "license", "antena_users", "goose_db_version"} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Fatalf("table %s was not created: %v", table, err)
		}
	}

	var clientIDCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('license') WHERE name = 'client_id'`).Scan(&clientIDCount); err != nil {
		t.Fatalf("failed to inspect license columns: %v", err)
	}
	if clientIDCount != 1 {
		t.Fatalf("license.client_id column count = %d, want 1", clientIDCount)
	}
}

func TestInsertLicense(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	db.SetMaxOpenConns(1)
	defer db.Close()

	ctx := context.Background()
	if err := clientstorage.InitRemoteSchema(ctx, db); err != nil {
		t.Fatalf("InitRemoteSchema() error = %v", err)
	}

	if err := clientstorage.InsertLicense(ctx, db, clientstorage.License{
		Nome:     "Hospital",
		ClientID: "hospitalsj",
		Validade: "2027-12-31",
		Status:   "active",
	}); err != nil {
		t.Fatalf("InsertLicense() error = %v", err)
	}

	var nome, clientID, status string
	err = db.QueryRow(`SELECT nome, client_id, status FROM license WHERE client_id = ?`, "hospitalsj").
		Scan(&nome, &clientID, &status)
	if err != nil {
		t.Fatalf("query license: %v", err)
	}
	if nome != "Hospital" || clientID != "hospitalsj" || status != "active" {
		t.Fatalf("license row = (%q, %q, %q)", nome, clientID, status)
	}
}
