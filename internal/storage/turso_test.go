package storage

import (
	"os"
	"path/filepath"
	"signal-admin/internal/migrations"
	"signal-admin/internal/models"
	"testing"
)

func TestClientesCRUDLocal(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := NewTursoDB(DBConfig{
		Mode:      "local",
		LocalPath: dbPath,
	})
	if err != nil {
		t.Fatalf("NewTursoDB: %v", err)
	}
	defer db.Close()
	defer os.Remove(dbPath)

	if err := migrations.Run(db.DB()); err != nil {
		t.Fatalf("migrations.Run: %v", err)
	}

	c := modelsCliente()
	if err := db.CreateCliente(c); err != nil {
		t.Fatalf("CreateCliente: %v", err)
	}

	got, err := db.GetCliente("cliente1")
	if err != nil {
		t.Fatalf("GetCliente: %v", err)
	}
	if got.Nome != "Cliente Teste" {
		t.Fatalf("Nome = %q, want Cliente Teste", got.Nome)
	}

	c.Nome = "Cliente Atualizado"
	if err := db.UpdateCliente(c); err != nil {
		t.Fatalf("UpdateCliente: %v", err)
	}

	list, total, err := db.ListClientesFilter("", "Atualizado", "", 10, 0)
	if err != nil {
		t.Fatalf("ListClientesFilter: %v", err)
	}
	if total != 1 || len(list) != 1 {
		t.Fatalf("list = %d total = %d, want 1/1", len(list), total)
	}

	if err := db.DeleteCliente("cliente1"); err != nil {
		t.Fatalf("DeleteCliente: %v", err)
	}

	has, err := db.HasUsers()
	if err != nil {
		t.Fatalf("HasUsers: %v", err)
	}
	if has {
		t.Fatal("HasUsers = true, want false")
	}

	if err := db.CreateUser("admin", "Admin", "secret123"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	user, err := db.Authenticate("admin", "secret123")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if user.Usuario != "admin" {
		t.Fatalf("Usuario = %q, want admin", user.Usuario)
	}
}

func modelsCliente() models.Cliente {
	return models.Cliente{
		ClienteID:      "cliente1",
		Nome:           "Cliente Teste",
		CNPJ:           "",
		Email:      "test@example.com",
		Telefone:   "",
		ValidUntil: "2027-12-31",
		Status:     "active",
		Observacao:     "obs",
	}
}
