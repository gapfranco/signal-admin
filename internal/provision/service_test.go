package provision_test

import (
	"log/slog"
	"os"
	"testing"

	"signal-admin/internal/models"
	"signal-admin/internal/provision"
)

func TestServiceEnabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	tests := []struct {
		name  string
		org   string
		token string
		want  bool
	}{
		{"both set", "org", "tok", true},
		{"org empty", "", "tok", false},
		{"token empty", "org", "", false},
		{"both empty", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := provision.NewService(tt.org, tt.token, logger)
			if got := s.Enabled(); got != tt.want {
				t.Fatalf("Enabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateFlashMessage(t *testing.T) {
	tests := []struct {
		name    string
		outcome provision.Outcome
		want    string
	}{
		{
			name:    "created",
			outcome: provision.Outcome{Status: provision.StatusCreated},
			want:    "Cliente criado e banco Turso provisionado.",
		},
		{
			name:    "already exists",
			outcome: provision.Outcome{Status: provision.StatusAlreadyExists},
			want:    "Cliente criado. Banco Turso já existia para este código.",
		},
		{
			name:    "failed create database",
			outcome: provision.Outcome{Status: provision.StatusFailed, Err: "criar banco: unauthorized"},
			want:    "Cliente criado, mas falha ao criar o banco no Turso. Detalhe: criar banco: unauthorized",
		},
		{
			name:    "failed verify database",
			outcome: provision.Outcome{Status: provision.StatusFailed, Err: "verificar banco: connection refused"},
			want:    "Cliente criado, mas falha ao verificar se o banco já existe no Turso. Detalhe: verificar banco: connection refused",
		},
		{
			name:    "failed create token",
			outcome: provision.Outcome{Status: provision.StatusFailed, Err: "criar token do banco: forbidden"},
			want:    "Cliente criado, mas falha ao gerar o token do banco. Detalhe: criar token do banco: forbidden",
		},
		{
			name:    "failed connect",
			outcome: provision.Outcome{Status: provision.StatusFailed, Err: "conectar ao banco criado (hospitalsj): ping remote: timeout"},
			want:    "Cliente criado, mas falha ao conectar ao banco criado. Detalhe: conectar ao banco criado (hospitalsj): ping remote: timeout",
		},
		{
			name:    "failed schema",
			outcome: provision.Outcome{Status: provision.StatusFailed, Err: "aplicar schema remoto: migration failed"},
			want:    "Cliente criado, mas falha ao aplicar o schema remoto. Detalhe: aplicar schema remoto: migration failed",
		},
		{
			name:    "failed license",
			outcome: provision.Outcome{Status: provision.StatusFailed, Err: "registrar licença: UNIQUE constraint"},
			want:    "Cliente criado, mas falha ao registrar a licença. Detalhe: registrar licença: UNIQUE constraint",
		},
		{
			name:    "failed config files",
			outcome: provision.Outcome{Status: provision.StatusFailed, Err: "escrever /tmp/hospitalsj-signal.conf: permission denied"},
			want:    "Cliente criado, mas falha ao gerar os arquivos de configuração. Detalhe: escrever /tmp/hospitalsj-signal.conf: permission denied",
		},
		{
			name:    "failed generic",
			outcome: provision.Outcome{Status: provision.StatusFailed, Err: "timeout"},
			want:    "Cliente criado, mas falha no provisionamento Turso. Detalhe: timeout",
		},
		{
			name:    "skipped",
			outcome: provision.Outcome{Status: provision.StatusSkipped},
			want:    "Cliente criado (provisionamento Turso não configurado).",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := provision.CreateFlashMessage(tt.outcome); got != tt.want {
				t.Fatalf("CreateFlashMessage() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeleteFlashMessage(t *testing.T) {
	if got := provision.DeleteFlashMessage("hospitalsj", true); got != `Cliente excluído. O banco Turso "hospitalsj" continua existente na nuvem.` {
		t.Fatalf("DeleteFlashMessage(exists) = %q", got)
	}
	if got := provision.DeleteFlashMessage("hospitalsj", false); got != "Cliente excluído." {
		t.Fatalf("DeleteFlashMessage(not exists) = %q", got)
	}
}

func TestProvisionClienteSkippedWhenDisabled(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	s := provision.NewService("", "", logger)

	outcome := s.ProvisionCliente(t.Context(), models.Cliente{
		ClienteID: "hospitalsj",
		Nome:      "Hospital",
	})
	if outcome.Status != provision.StatusSkipped {
		t.Fatalf("Status = %v, want StatusSkipped", outcome.Status)
	}
}
