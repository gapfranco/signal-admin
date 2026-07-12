package provision_test

import (
	"strings"
	"testing"

	"signal-admin/internal/provision"
)

func TestValidateOptions(t *testing.T) {
	tests := []struct {
		name    string
		client  string
		db      string
		limit   string
		org     string
		token   string
		wantErr string
	}{
		{
			name:    "client empty",
			db:      "hospitalsj",
			org:     "myorg",
			token:   "tok",
			wantErr: "nome do cliente é obrigatório",
		},
		{
			name:    "db empty",
			client:  "Hospital",
			org:     "myorg",
			token:   "tok",
			wantErr: "código do cliente é obrigatório",
		},
		{
			name:    "db uppercase",
			client:  "Hospital",
			db:      "HospitalSJ",
			org:     "myorg",
			token:   "tok",
			wantErr: "slug",
		},
		{
			name:    "limit invalid format",
			client:  "Hospital",
			db:      "hospitalsj",
			limit:   "31/12/2027",
			org:     "myorg",
			token:   "tok",
			wantErr: "AAAA-MM-DD",
		},
		{
			name:    "org missing",
			client:  "Hospital",
			db:      "hospitalsj",
			token:   "tok",
			wantErr: "TURSO_ORG",
		},
		{
			name:    "token missing",
			client:  "Hospital",
			db:      "hospitalsj",
			org:     "myorg",
			wantErr: "TURSO_TOKEN",
		},
		{
			name:   "valid without limit",
			client: "Hospital São João",
			db:     "hospitalsj",
			org:    "myorg",
			token:  "tok",
		},
		{
			name:   "valid with limit",
			client: "Condomínio Alfa",
			db:     "alfacond",
			limit:  "2027-12-31",
			org:    "myorg",
			token:  "tok",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provision.ValidateOptions(tt.client, tt.db, tt.limit, tt.org, tt.token)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("ValidateOptions() error = nil, want %q", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ValidateOptions() error = %q, want substring %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ValidateOptions() error = %v", err)
			}
			if got.Client != strings.TrimSpace(tt.client) {
				t.Fatalf("Client = %q, want %q", got.Client, tt.client)
			}
			if got.DB != tt.db {
				t.Fatalf("DB = %q, want %q", got.DB, tt.db)
			}
		})
	}
}

func TestLibSQLURL(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"mydb-org.aws-us-east-1.turso.io", "libsql://mydb-org.aws-us-east-1.turso.io"},
		{"libsql://mydb-org.aws-us-east-1.turso.io", "libsql://mydb-org.aws-us-east-1.turso.io"},
	}
	for _, tt := range tests {
		if got := provision.LibSQLURL(tt.in); got != tt.want {
			t.Fatalf("LibSQLURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
