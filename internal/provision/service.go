package provision

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"signal-admin/internal/models"
)

type Status int

const (
	StatusCreated Status = iota
	StatusAlreadyExists
	StatusFailed
	StatusSkipped
)

type Outcome struct {
	Status Status
	Err    string
	Result Result
}

type Service struct {
	org    string
	token  string
	logger *slog.Logger
}

func NewService(org, token string, logger *slog.Logger) *Service {
	return &Service{org: org, token: token, logger: logger}
}

func (s *Service) Enabled() bool {
	return s.org != "" && s.token != ""
}

func (s *Service) ProvisionCliente(ctx context.Context, cliente models.Cliente) Outcome {
	if !s.Enabled() {
		s.logger.Warn("provision skipped: TURSO_ORG or TURSO_TOKEN not configured")
		return Outcome{Status: StatusSkipped}
	}

	opts, err := ValidateOptions(cliente.Nome, cliente.ClienteID, cliente.ValidUntil, s.org, s.token)
	if err != nil {
		s.logger.Error("provision validation failed", "cliente_id", cliente.ClienteID, "error", err)
		return Outcome{Status: StatusFailed, Err: err.Error()}
	}

	result, err := Run(ctx, opts)
	if errors.Is(err, ErrDatabaseExists) {
		s.logger.Warn("turso database already exists", "cliente_id", cliente.ClienteID)
		return Outcome{Status: StatusAlreadyExists}
	}
	if err != nil {
		s.logger.Error("provision failed", "cliente_id", cliente.ClienteID, "error", err)
		return Outcome{Status: StatusFailed, Err: err.Error()}
	}

	s.logger.Info("turso database provisioned",
		"cliente_id", result.DBName,
		"signal_conf", result.Configs.Signal,
		"antena_conf", result.Configs.Antena,
	)
	return Outcome{Status: StatusCreated, Result: result}
}

func (s *Service) ClienteDBExists(ctx context.Context, clienteID string) (bool, error) {
	if !s.Enabled() {
		return false, nil
	}

	api, err := NewTursoAPI(s.org, s.token)
	if err != nil {
		return false, err
	}
	return api.DatabaseExists(ctx, clienteID)
}

func CreateFlashMessage(outcome Outcome) string {
	switch outcome.Status {
	case StatusCreated:
		return "Cliente criado e banco Turso provisionado."
	case StatusAlreadyExists:
		return "Cliente criado. Banco Turso já existia para este código."
	case StatusFailed:
		return formatProvisionFailure(outcome.Err)
	case StatusSkipped:
		return "Cliente criado (provisionamento Turso não configurado)."
	default:
		return "Cliente criado com sucesso."
	}
}

func formatProvisionFailure(errDetail string) string {
	friendly := "falha no provisionamento Turso"
	switch {
	case strings.Contains(errDetail, "verificar banco"):
		friendly = "falha ao verificar se o banco já existe no Turso"
	case strings.Contains(errDetail, "criar banco"):
		friendly = "falha ao criar o banco no Turso"
	case strings.Contains(errDetail, "criar token"):
		friendly = "falha ao gerar o token do banco"
	case strings.Contains(errDetail, "conectar ao banco"):
		friendly = "falha ao conectar ao banco criado"
	case strings.Contains(errDetail, "aplicar schema"):
		friendly = "falha ao aplicar o schema remoto"
	case strings.Contains(errDetail, "registrar licença"):
		friendly = "falha ao registrar a licença"
	case strings.Contains(errDetail, "escrever"), strings.Contains(errDetail, "obter diretório"):
		friendly = "falha ao gerar os arquivos de configuração"
	}

	if strings.TrimSpace(errDetail) == "" {
		return "Cliente criado, mas " + friendly + "."
	}
	return "Cliente criado, mas " + friendly + ". Detalhe: " + errDetail
}

func DeleteFlashMessage(clienteID string, dbExists bool) string {
	if dbExists {
		return "Cliente excluído. O banco Turso \"" + clienteID + "\" continua existente na nuvem."
	}
	return "Cliente excluído."
}
